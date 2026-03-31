# Go Claude Code

Claude Code 的 Go 实现 - 一个强大的 AI 驱动命令行工具，将 Claude 的能力带到你的终端。

## 功能特性

- **交互式对话**：在交互式 REPL 中与 Claude 聊天
- **文件操作**：读取、写入和编辑文件
- **终端命令**：直接执行 shell 命令
- **工具系统**：可扩展的工具架构
- **MCP 支持**：连接模型上下文协议（MCP）服务器
- **对话历史**：持久化和管理对话历史记录

## 安装

### 从源码构建

```bash
git clone https://github.com/ruizhehou/go-claude-code.git
cd go-claude-code
make build
```

二进制文件将生成在 `bin/claude`。

### 安装到 GOPATH

```bash
make install
```

## 配置

### 环境变量

- `CLAUDE_PROVIDER`：使用的 API 提供商 - `anthropic`（默认）或 `volcengine`
- `ANTHROPIC_API_KEY`：你的 Anthropic API 密钥（使用 Anthropic 时必需）
- `CLAUDE_MODEL`：使用的模型（Anthropic 默认：`claude-sonnet-4-20250514`）
- `CLAUDE_MAX_TOKENS`：响应的最大 token 数（默认：4096）
- `CLAUDE_TEMPERATURE`：响应温度（默认：0.0）
- `CLAUDE_ALLOWED_PATHS`：允许访问的文件路径（冒号分隔）
- `CLAUDE_SHELL_ENABLED`：启用/禁用 shell 工具（默认：true）
- `VOLCENGINE_API_KEY`：你的火山引擎 API 密钥（使用火山引擎时必需）
- `VOLCENGINE_MODEL`：火山引擎模型（默认：`ep-20241125174025-l8jx8`）
- `VOLCENGINE_BASE_URL`：火山引擎 API 基础 URL（默认：`https://ark.cn-beijing.volces.com/api/v3`）

### 配置文件

在 `~/.claude-code/config.json` 创建配置文件：

**使用 Anthropic 提供商（默认）：**

```json
{
  "provider": "anthropic",
  "anthropic": {
    "apiKey": "你的-api-key",
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

**使用火山引擎提供商：**

```json
{
  "provider": "volcengine",
  "volcengine": {
    "apiKey": "你的火山引擎-api-key",
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

### MCP 服务器配置

在 `~/.claude-code/servers.json` 创建 MCP 服务器配置：

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

## 使用方法

### 基本对话

```bash
# 使用 Anthropic（默认）
export ANTHROPIC_API_KEY=sk-ant-xxx
./bin/claude

# 使用火山引擎
export CLAUDE_PROVIDER=volcengine
export VOLCENGINE_API_KEY=你的火山引擎-api-key
./bin/claude
```

### 命令

- `exit`、`quit` - 退出程序
- `clear` - 清除对话历史
- `help` - 显示帮助信息
- `tools` - 列出可用工具

### 内置工具

#### Bash 工具

执行 shell 命令：

```
claude> 列出当前目录的文件
```

#### 文件读取工具

读取文件内容：

```
claude> 读取 /path/to/file.txt 文件的内容
```

#### 文件写入工具

写入内容到文件：

```
claude> 创建一个新文件 /path/to/file.txt，内容是 "Hello, World!"
```

#### 文件编辑工具

编辑现有文件：

```
claude> 把 /path/to/file.txt 中的 "旧文本" 替换成 "新文本"
```

## 项目结构

```
go-claude-code/
├── cmd/claude/          # CLI 入口点
├── pkg/
│   ├── api/            # API 客户端（Anthropic、火山引擎）
│   ├── cli/            # REPL 实现
│   ├── tools/          # 工具系统
│   │   ├── builtin/    # 内置工具
│   │   └── mcp/        # MCP 客户端
│   ├── config/         # 配置管理
│   ├── conversation/   # 对话历史
│   └── utils/          # 工具函数
├── internal/
│   └── protocol/
│       └── mcp/        # MCP 协议定义
└── config/             # 示例配置文件
```

## 开发

### 构建

```bash
make build
```

### 运行

```bash
make run
```

### 测试

```bash
make test
```

### 格式化代码

```bash
make fmt
```

## 快速开始

```bash
# 1. 设置 API 密钥（选择一个提供商）
# 使用 Anthropic（默认）：
export ANTHROPIC_API_KEY=sk-ant-xxx

# 或使用火山引擎：
export CLAUDE_PROVIDER=volcengine
export VOLCENGINE_API_KEY=你的火山引擎-api-key

# 2. 运行程序
./bin/claude

# 3. 开始对话
claude> 帮我写一个 Hello World 程序
```

## 常见问题

### 如何获取 API 密钥？

**Anthropic**：访问 [Anthropic 控制台](https://console.anthropic.com/) 注册账号并获取 API 密钥。

**火山引擎**：访问 [火山引擎控制台](https://console.volcengine.com/) 注册账号并获取 API 密钥。

### 支持哪些模型？

**Anthropic 提供商：**
目前支持所有 Anthropic Claude 模型，包括：
- `claude-sonnet-4-20250514`（默认）
- `claude-opus-4-20250514`
- `claude-haiku-4-20250514`

**火山引擎提供商：**
支持火山引擎的豆包模型，如：
- `ep-20241125174025-l8jx8`（默认）
- 其他自定义推理端点

### 如何连接 MCP 服务器？

1. 创建 `~/.claude-code/servers.json` 配置文件
2. 添加你需要的 MCP 服务器配置
3. 重新启动程序，工具会自动加载

### 如何切换提供商？

设置 `CLAUDE_PROVIDER` 环境变量或在配置文件中配置：

```bash
# 使用 Anthropic（默认）
export CLAUDE_PROVIDER=anthropic

# 使用火山引擎
export CLAUDE_PROVIDER=volcengine
```

## 许可证

MIT License

## 贡献

欢迎贡献！请随时提交 Pull Request。

## 相关链接

- [Anthropic API 文档](https://docs.anthropic.com/)
- [MCP 协议规范](https://modelcontextprotocol.io/)
- [Claude Code 官方文档](https://claude.ai/code)
