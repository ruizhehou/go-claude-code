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

- `ANTHROPIC_API_KEY`: Your Anthropic API key (required)
- `CLAUDE_MODEL`: Model to use (default: `claude-sonnet-4-20250514`)
- `CLAUDE_MAX_TOKENS`: Maximum tokens for responses (default: 4096)
- `CLAUDE_TEMPERATURE`: Temperature for responses (default: 0.0)
- `CLAUDE_ALLOWED_PATHS`: Colon-separated list of allowed file paths
- `CLAUDE_SHELL_ENABLED`: Enable/disable shell tool (default: true)

### Configuration File

Create a configuration file at `~/.claude-code/config.json`:

```json
{
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
│   ├── api/            # Anthropic API client
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
