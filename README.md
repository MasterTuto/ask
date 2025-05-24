# ask - Universal LLM CLI Tool 🚀

A powerful command-line interface for interacting with multiple LLM providers through a single, unified tool. No more juggling between different apps or memorizing various API syntaxes.

![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen?style=for-the-badge)

## ✨ Features

- **Multiple Providers**: Support for Claude, OpenAI, Google Gemini, Cohere, and local models
- **Simple Syntax**: Intuitive commands that just make sense
- **Secure Storage**: API keys stored securely with proper file permissions
- **Fast & Lightweight**: Built in Go for maximum performance
- **Local Model Support**: Run models locally via Ollama integration
- **Zero Config**: Works out of the box with minimal setup

## 🚀 Quick Start

```bash
# Install
go install github.com/yourusername/ask@latest

# Add your first API
ask add api:claude

# Run your first prompt
ask api:claude "Write a hello world in Go"
```

## 📦 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/ask.git
cd ask

# Install dependencies
go mod tidy

# Build
go build -o ask

# Move to PATH
sudo mv ask /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/yourusername/ask@latest
```

### Prerequisites

- Go 1.20 or higher
- [Ollama](https://ollama.ai) (for local models)

## 🎯 Usage

### Basic Commands

```bash
# Run a prompt
ask <provider> "<prompt>"

# Add a new API/model
ask add <provider>

# List all configured APIs
ask list

# Remove an API
ask remove <provider>
```

### Examples

```bash
# API Models
ask api:claude "Generate a REST API in Python"
ask api:gpt-4 "Explain quantum computing"
ask api:gemini-pro "Write unit tests for this function"
ask api:cohere "Summarize this article"

# Local Models
ask local:llama3-8b "Write a poem about coding"
ask local:mistral-7b "Debug this JavaScript code"
ask local:deepseek-r1-8b "Create a dockerfile"

# Adding APIs
ask add api:claude              # Adds Claude with default model
ask add api:claude-opus          # Adds Claude with Opus model
ask add api:gpt-4o              # Adds GPT-4o specifically
ask add local:codellama-13b     # Adds local model

# Management
ask list                        # Show all configured APIs
ask remove api:claude           # Remove an API
```

## 🤖 Supported Providers

### API Providers

| Provider | Models | Command Example |
|----------|--------|-----------------|
| **Claude** | Claude 3.5 Sonnet, Claude 3 Opus, Claude 3 Haiku | `ask api:claude` |
| **OpenAI** | GPT-4, GPT-4 Turbo, GPT-4o, GPT-3.5 | `ask api:gpt-4` |
| **Google Gemini** | Gemini 1.5 Pro, Gemini 1.5 Flash | `ask api:gemini` |
| **Cohere** | Command R+, Command R | `ask api:cohere` |

### Local Models (via Ollama)

- Llama 3 (8B, 70B)
- Mistral (7B)
- DeepSeek R1 (8B)
- CodeLlama
- And any model supported by Ollama!

## ⚙️ Configuration

Configuration is stored in `~/.ask/config.json`:

```json
{
  "apis": {
    "api:claude": {
      "provider": "claude",
      "api_key": "sk-ant-...",
      "base_url": "https://api.anthropic.com/v1",
      "model": "claude-3-5-sonnet-20241022"
    },
    "local:llama3": {
      "provider": "local",
      "model": "llama3"
    }
  }
}
```

## 🔐 Security

- API keys are stored locally in `~/.ask/config.json`
- File permissions are set to `0600` (owner read/write only)
- Keys are never logged or exposed in terminal output
- Password-style input (hidden) when entering API keys

## 🛠️ Development

### Building from Source

```bash
# Clone and enter directory
git clone https://github.com/yourusername/ask.git
cd ask

# Get dependencies
go mod tidy

# Run tests
go test ./...

# Build
go build -o ask

# Run locally
./ask api:claude "Hello, world!"
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI management
- Local model support via [Ollama](https://ollama.ai)
- Inspired by the need for a unified LLM interface

## 🤔 FAQ

**Q: Is my API key secure?**  
A: Yes, API keys are stored locally with restricted file permissions and never transmitted except to the respective API providers.

**Q: Can I use custom API endpoints?**  
A: Yes, you can modify the base URLs in the config file for custom endpoints.

**Q: How do I update models?**  
A: Simply remove and re-add the API with the new model, or edit the config file directly.

**Q: Does it support streaming responses?**  
A: Currently, responses are displayed after completion. Streaming support is planned for v2.

## 🚧 Roadmap

- [ ] Streaming responses
- [ ] Conversation history
- [ ] Multiple message support
- [ ] File input support
- [ ] Custom system prompts
- [ ] Export conversations
- [ ] Plugin system

---

Made with ❤️ by developers, for developers. Stop switching between apps, start shipping faster.
