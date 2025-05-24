package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

// Config structure to store API keys and settings
type Config struct {
	APIs map[string]APIConfig `json:"apis"`
}

type APIConfig struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url,omitempty"`
	Model    string `json:"model"`
}

// Supported providers
const (
	ProviderClaude   = "claude"
	ProviderOpenAI   = "openai"
	ProviderGemini   = "gemini"
	ProviderCohere   = "cohere"
	ProviderLocal    = "local"
)

// Model mappings
var modelMappings = map[string]string{
	// Claude models
	"claude":        "claude-3-5-sonnet-20241022",
	"claude-3":      "claude-3-5-sonnet-20241022",
	"claude-3.5":    "claude-3-5-sonnet-20241022",
	"claude-opus":   "claude-3-opus-20240229",
	"claude-sonnet": "claude-3-5-sonnet-20241022",
	"claude-haiku":  "claude-3-haiku-20240307",
	
	// OpenAI models
	"gpt-4":        "gpt-4-turbo-preview",
	"gpt-4-turbo":  "gpt-4-turbo-preview",
	"gpt-3.5":      "gpt-3.5-turbo",
	"gpt-4o":       "gpt-4o",
	"gpt-4o-mini":  "gpt-4o-mini",
	
	// Gemini models
	"gemini":        "gemini-1.5-pro",
	"gemini-pro":    "gemini-1.5-pro",
	"gemini-flash":  "gemini-1.5-flash",
	
	// Cohere models
	"cohere":        "command-r-plus",
	"command":       "command-r-plus",
	"command-light": "command-r",
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	config := loadConfig()

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ask add <api:provider-model|local:model>")
			os.Exit(1)
		}
		addAPI(config, os.Args[2])
	case "list":
		listAPIs(config)
	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ask remove <api-name>")
			os.Exit(1)
		}
		removeAPI(config, os.Args[2])
	default:
		// Assume it's a prompt command
		if len(os.Args) < 3 {
			fmt.Println("Usage: ask <api:provider|local:model> \"<prompt>\"")
			os.Exit(1)
		}
		runPrompt(config, os.Args[1], strings.Join(os.Args[2:], " "))
	}
}

func printUsage() {
	fmt.Println(`ask - CLI tool for interacting with LLMs

Usage:
  ask <api:provider|local:model> "<prompt>"    Run a prompt
  ask add <api:provider-model|local:model>     Add a new API/model
  ask list                                      List configured APIs
  ask remove <api-name>                         Remove an API

Examples:
  ask api:claude "generate an index.ts file"
  ask api:gpt-4 "explain quantum computing"
  ask local:deepseek-r1-8b "write a poem"
  ask add api:claude-opus
  ask add local:llama3-8b

Supported API providers:
  - claude (Claude 3/3.5 models)
  - openai (GPT-3.5, GPT-4, GPT-4o)
  - gemini (Gemini Pro, Flash)
  - cohere (Command R/R+)

Supported local models:
  - deepseek-r1-8b
  - llama3-8b
  - mistral-7b
  - And any model supported by ollama`)
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ask", "config.json")
}

func loadConfig() *Config {
	configPath := getConfigPath()
	config := &Config{APIs: make(map[string]APIConfig)}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config
	}

	json.Unmarshal(data, config)
	return config
}

func saveConfig(config *Config) error {
	configPath := getConfigPath()
	os.MkdirAll(filepath.Dir(configPath), 0755)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func addAPI(config *Config, apiSpec string) {
	parts := strings.Split(apiSpec, ":")
	if len(parts) != 2 {
		fmt.Println("Invalid format. Use api:provider-model or local:model")
		os.Exit(1)
	}

	apiType := parts[0]
	providerModel := parts[1]

	if apiType == "local" {
		// Local model
		config.APIs[apiSpec] = APIConfig{
			Provider: ProviderLocal,
			Model:    providerModel,
		}
		saveConfig(config)
		fmt.Printf("Added local model: %s\n", providerModel)
		return
	}

	// API model
	provider := ""
	model := ""

	// Parse provider and model
	if strings.Contains(providerModel, "-") {
		// Check if it's a specific model
		if mappedModel, ok := modelMappings[providerModel]; ok {
			model = mappedModel
			// Determine provider from model name
			if strings.HasPrefix(providerModel, "claude") {
				provider = ProviderClaude
			} else if strings.HasPrefix(providerModel, "gpt") {
				provider = ProviderOpenAI
			} else if strings.HasPrefix(providerModel, "gemini") {
				provider = ProviderGemini
			} else if strings.HasPrefix(providerModel, "command") || strings.HasPrefix(providerModel, "cohere") {
				provider = ProviderCohere
			}
		} else {
			// Try to parse as provider-model
			parts := strings.SplitN(providerModel, "-", 2)
			provider = parts[0]
			if len(parts) > 1 {
				model = providerModel
			}
		}
	} else {
		// Just provider name
		provider = providerModel
		// Use default model for provider
		switch provider {
		case ProviderClaude:
			model = "claude-3-5-sonnet-20241022"
		case "openai":
			model = "gpt-4-turbo-preview"
		case "gemini":
			model = "gemini-1.5-pro"
		case "cohere":
			model = "command-r-plus"
		}
	}

	// Prompt for API key
	fmt.Printf("Enter API key for %s: ", provider)
	apiKey, err := readPassword()
	if err != nil {
		fmt.Println("\nError reading API key:", err)
		os.Exit(1)
	}

	baseURL := ""
	switch provider {
	case ProviderClaude:
		baseURL = "https://api.anthropic.com/v1"
	case ProviderOpenAI:
		baseURL = "https://api.openai.com/v1"
	case ProviderGemini:
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	case ProviderCohere:
		baseURL = "https://api.cohere.ai/v1"
	}

	config.APIs[apiSpec] = APIConfig{
		Provider: provider,
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Model:    model,
	}

	saveConfig(config)
	fmt.Printf("\nAdded API: %s (provider: %s, model: %s)\n", apiSpec, provider, model)
}

func readPassword() (string, error) {
	fd := int(syscall.Stdin)
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	var password []byte
	reader := bufio.NewReader(os.Stdin)

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}

		if b == '\n' || b == '\r' {
			break
		} else if b == 127 || b == 8 { // Backspace
			if len(password) > 0 {
				password = password[:len(password)-1]
			}
		} else if b >= 32 && b <= 126 { // Printable characters
			password = append(password, b)
		}
	}

	return string(password), nil
}

func listAPIs(config *Config) {
	if len(config.APIs) == 0 {
		fmt.Println("No APIs configured. Use 'ask add' to add one.")
		return
	}

	fmt.Println("Configured APIs:")
	for name, api := range config.APIs {
		fmt.Printf("  %s (provider: %s, model: %s)\n", name, api.Provider, api.Model)
	}
}

func removeAPI(config *Config, apiName string) {
	if _, exists := config.APIs[apiName]; !exists {
		fmt.Printf("API '%s' not found\n", apiName)
		return
	}

	delete(config.APIs, apiName)
	saveConfig(config)
	fmt.Printf("Removed API: %s\n", apiName)
}

func runPrompt(config *Config, apiSpec string, prompt string) {
	apiConfig, exists := config.APIs[apiSpec]
	if !exists {
		fmt.Printf("API '%s' not configured. Use 'ask add %s' to add it.\n", apiSpec, apiSpec)
		os.Exit(1)
	}

	if apiConfig.Provider == ProviderLocal {
		runLocalModel(apiConfig.Model, prompt)
		return
	}

	switch apiConfig.Provider {
	case ProviderClaude:
		runClaude(apiConfig, prompt)
	case ProviderOpenAI:
		runOpenAI(apiConfig, prompt)
	case ProviderGemini:
		runGemini(apiConfig, prompt)
	case ProviderCohere:
		runCohere(apiConfig, prompt)
	default:
		fmt.Printf("Unknown provider: %s\n", apiConfig.Provider)
		os.Exit(1)
	}
}

func runClaude(config APIConfig, prompt string) {
	url := config.BaseURL + "/messages"
	
	payload := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 4096,
	}

	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n%s\n", resp.Status, string(body))
		os.Exit(1)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
		if text, ok := content[0].(map[string]interface{})["text"].(string); ok {
			fmt.Println(text)
		}
	}
}

func runOpenAI(config APIConfig, prompt string) {
	url := config.BaseURL + "/chat/completions"
	
	payload := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n%s\n", resp.Status, string(body))
		os.Exit(1)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{}); ok {
			if content, ok := message["content"].(string); ok {
				fmt.Println(content)
			}
		}
	}
}

func runGemini(config APIConfig, prompt string) {
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", config.BaseURL, config.Model, config.APIKey)
	
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n%s\n", resp.Status, string(body))
		os.Exit(1)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if candidates, ok := result["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if content, ok := candidates[0].(map[string]interface{})["content"].(map[string]interface{}); ok {
			if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
				if text, ok := parts[0].(map[string]interface{})["text"].(string); ok {
					fmt.Println(text)
				}
			}
		}
	}
}

func runCohere(config APIConfig, prompt string) {
	url := config.BaseURL + "/chat"
	
	payload := map[string]interface{}{
		"model":   config.Model,
		"message": prompt,
	}

	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n%s\n", resp.Status, string(body))
		os.Exit(1)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if text, ok := result["text"].(string); ok {
		fmt.Println(text)
	}
}

func runLocalModel(model string, prompt string) {
	// Check if ollama is installed
	_, err := exec.LookPath("ollama")
	if err != nil {
		fmt.Println("Error: ollama not found. Please install ollama to use local models.")
		fmt.Println("Visit: https://ollama.ai")
		os.Exit(1)
	}

	// Run ollama with the model
	cmd := exec.Command("ollama", "run", model, prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running model %s: %v\n", model, err)
		os.Exit(1)
	}
}
