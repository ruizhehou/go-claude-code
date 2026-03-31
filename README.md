# Go Claude Code

A Go implementation of Claude Code - an AI-powered CLI tool that brings the power of Claude to your terminal.

[中文文档](README_CN.md) | [English](README.md)

## Features

- **Interactive Chat**: Chat with Claude in an interactive REPL
- **File Operations**: Read, write, and edit files
- **Terminal Commands**: Execute shell commands directly
- **Tool System**: Extensible tool architecture
- **MCP Support**: Connect to Model Context Protocol servers
- **Conversation History**: Persist and manage conversation history

## Installation

### Build from source

```bash
git clone https://github.com/houruizhe/go-claude-code.git
cd go-claude-code
make build
```

The binary will be created at `bin/claude`.

### Install to GOPATH

```bash
make install
```

## Configuration

### Environment Variables

- `CLAUDE_PROVIDER`: API provider to use - `anthropic` (default) or `volcengine`
- `ANTHROPIC_API_KEY`: Your Anthropic API key (required when using Anthropic provider)
- `CLAUDE_MODEL`: Model to use (default: `claude-sonnet-4-20250514` for Anthropic)
- `CLAUDE_MAX_TOKENS`: Maximum tokens for responses (default: 4096)
- `CLAUDE_TEMPERATURE`: Temperature for responses (default: 0.0)
- `CLAUDE_ALLOWED_PATHS`: Colon-separated list of allowed file paths
- `CLAUDE_SHELL_ENABLED`: Enable/disable shell tool (default: true)
- `VOLCENGINE_API_KEY`: Your Volcengine API key (required when using Volcengine provider)
- `VOLCENGINE_MODEL`: Volcengine model to use (default: `ep-20241125174025-l8jx8`)
- `VOLCENGINE_BASE_URL`: Volcengine API base URL (default: `https://ark.cn-beijing.volces.com/api/v3`)

### Configuration File

Create a configuration file at `~/.claude-code/config.json`:

**Using Anthropic provider (default):**

```json
{
  "provider": "anthropic",
  "anthropic": {
    "apiKey": "your-api-key",
    "model": "claude-sonnet-4-20250514",
    "maxTokens": 4096,
    "temperature": 0.0
  },
  "cli": {
    "prompt": "claude> ",
    "historyFile": "~/.claude-code/history.json",
    "logLevel": "info"
  },
  "tools": {
    "allowedPaths": ["/home/user/projects"],
    "shell": {
      "enabled": true,
      "timeout": "2m"
    }
  }
}
```

**Using Volcengine provider:**

```json
{
  "provider": "volcengine",
  "volcengine": {
    "apiKey": "your-volcengine-api-key",
    "model": "ep-20241125174025-l8jx8",
    "maxTokens": 4096,
    "temperature": 0.0,
    "baseUrl": "https://ark.cn-beijing.volces.com/api/v3"
  },
  "cli": {
    "prompt": "claude> ",
    "historyFile": "~/.claude-code/history.json",
    "logLevel": "info"
  },
  "tools": {
    "allowedPaths": ["/home/user/projects"],
    "shell": {
      "enabled": true,
      "timeout": "2m"
    }
  }
}
```

### MCP Servers Configuration

Create MCP servers configuration at `~/.claude-code/servers.json`:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed"],
      "enabled": true
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "enabled": true
    }
  }
}
```

## Usage

### Basic Chat

```bash
# Using Anthropic (default)
export ANTHROPIC_API_KEY=sk-ant-xxx
./bin/claude

# Using Volcengine
export CLAUDE_PROVIDER=volcengine
export VOLCENGINE_API_KEY=your-volcengine-api-key
./bin/claude
```

### Commands

- `exit`, `quit` - Exit the program
- `clear` - Clear conversation history
- `help` - Show help message
- `tools` - List available tools

### Built-in Tools

#### Bash Tool

Execute shell commands:

```
claude> Run ls -la
```

#### File Read Tool

Read file contents:

```
claude> Read the file at /path/to/file.txt
```

#### File Write Tool

Write content to a file:

```
claude> Create a new file at /path/to/file.txt with content "Hello, World!"
```

#### File Edit Tool

Edit an existing file:

```
claude> Replace "old text" with "new text" in /path/to/file.txt
```

## Project Structure

```
go-claude-code/
├── cmd/claude/          # CLI entry point
├── pkg/
│   ├── api/            # API clients (Anthropic, Volcengine)
│   ├── cli/            # REPL implementation
│   ├── tools/          # Tool system
│   │   ├── builtin/    # Built-in tools
│   │   └── mcp/        # MCP client
│   ├── config/         # Configuration management
│   ├── conversation/   # Conversation history
│   └── utils/          # Utilities
├── internal/
│   └── protocol/
│       └── mcp/        # MCP protocol definitions
└── config/             # Example configurations
```

## Development

### Build

```bash
make build
```

### Run

```bash
make run
```

### Test

```bash
make test
```

### Format

```bash
make fmt
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## FAQ

### How to get an API key?

For Anthropic: Visit the [Anthropic Console](https://console.anthropic.com/) to sign up and get your API key.

For Volcengine: Visit the [Volcengine Console](https://console.volcengine.com/) to sign up and get your API key.

### Which models are supported?

**Anthropic provider:**
- `claude-sonnet-4-20250514` (default)
- `claude-opus-4-20250514`
- `claude-haiku-4-20250514`

**Volcengine provider:**
- `ep-20241125174025-l8jx8` (default)
- Other custom inference endpoints

### How to switch between providers?

Set the `CLAUDE_PROVIDER` environment variable or configure it in your config file:

```bash
# Using Anthropic (default)
export CLAUDE_PROVIDER=anthropic

# Using Volcengine
export CLAUDE_PROVIDER=volcengine
```
