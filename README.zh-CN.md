<p align="center">
  <img src="assets/logo.png" alt="Claude Code Go Logo" width="200">
</p>

<h1 align="center">Claude Code Go</h1>

<p align="center">
  <strong>🤖 Claude Code 的 Go 语言复刻版 — 终端里的 AI 编程助手</strong>
</p>

<p align="center">
  <a href="https://golang.org/dl/"><img src="https://img.shields.io/badge/go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go 版本"></a>
  <a href="https://goreportcard.com/report/github.com/808cn163/claude-code-go"><img src="https://goreportcard.com/badge/github.com/808cn163/claude-code-go?style=flat-square" alt="Go Report Card"></a>
  <a href="https://codecov.io/gh/808cn163/claude-code-go"><img src="https://codecov.io/gh/808cn163/claude-code-go/branch/main/graph/badge.svg?style=flat-square" alt="代码覆盖率"></a>
  <a href="https://pkg.go.dev/github.com/808cn163/claude-code-go"><img src="https://pkg.go.dev/badge/github.com/808cn163/claude-code-go.svg" alt="Go Reference"></a>
  <a href="https://github.com/808cn163/claude-code-go/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/808cn163/claude-code-go/ci.yml?branch=main&style=flat-square&logo=github&label=CI" alt="CI"></a>
  <a href="https://github.com/808cn163/claude-code-go/releases"><img src="https://img.shields.io/github/v/release/808cn163/claude-code-go?style=flat-square&logo=github" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="License"></a>
  <a href="https://github.com/808cn163/claude-code-go/pulls"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen?style=flat-square" alt="PRs Welcome"></a>
</p>

<p align="center">
  <a href="README.md">English</a> •
  <a href="README.zh-CN.md">简体中文</a> •
  <a href="README.ja.md">日本語</a> •
  <a href="README.ko.md">한국어</a> •
  <a href="README.es.md">Español</a> •
  <a href="README.fr.md">Français</a>
</p>

---

<p align="center">
  <img src="assets/image.png" alt="Claude Code Go - 交互式会话" width="800">
  <br>
  <em>交互式 TUI 界面，支持文件读取和实时思考状态显示</em>
</p>

<p align="center">
  <img src="assets/image1.png" alt="Claude Code Go - 项目分析" width="800">
  <br>
  <em>全面的项目分析报告，包含架构层级详情</em>
</p>

---

## 这是什么

本项目是对 [Claude Code](https://claude.ai/code)（Anthropic 官方 TypeScript CLI）**完整功能的 Go 语言复刻**，逐模块对照原版源码实现，覆盖 TUI 界面、工具调用、权限系统、多 Agent 协调、MCP 协议、会话管理等全部核心功能。它理解你的代码库、执行工具，通过自然对话帮你编写、评审与重构代码。

> **🤖 全程由 AI Agent 构建——零人工生产代码。** 初始构建（2026 年 4 月）由 9 个 Claude AI Agent 用 9 天协作完成；2026 年 8 月起由单个主 Claude 会话对照行为对等台账维护项目，人类负责人评审与合并。[完整故事 ↓](#零人工代码的构建故事)

## ❤️ 赞助商

> [想出现在这里？](https://github.com/tunsuy)

<table>
  <tr>
    <td width="180"><a href="https://www.orcarouter.ai/ref/ref_49eee7ba2e9927450075"><img src="assets/sponsors/orcarouter.png" alt="OrcaRouter" width="150"></a></td>
    <td><strong><a href="https://www.orcarouter.ai/ref/ref_49eee7ba2e9927450075">OrcaRouter</a></strong> —— 一个网关，所有模型。OrcaRouter 把每个请求在 200+ 个模型（前沿与开源）中路由到最合适的一个，价格与各提供商官方价格一致，并提供故障转移与请求日志。它兼容 OpenAI API，可直接接入 Claude Code Go——具体配置见 <a href="#api-提供商">API 提供商</a> 一节。</td>
  </tr>
</table>

## 功能特性

- **交互式 TUI** — 基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 构建的全功能终端界面，支持深色/浅色主题
- **智能工具调用** — 34 个内置工具（文件、命令、搜索、Web、任务、子 Agent 等），所有操作均经过 9 层权限流水线审批
- **CLI 子命令** — `claude mcp …`、`claude plugin …`、`claude auth …`、`claude doctor`、`claude update`；`mcp` 与 `plugin` 两组输出与 TypeScript 原版**字节级一致**
- **插件管理** — 从本地 marketplace 安装/启用/停用插件，带完整 manifest 校验（`claude plugin validate`）
- **多 Agent 协作** — 可启动后台子 Agent 并行处理任务
- **MCP 支持** — 通过 [Model Context Protocol](https://modelcontextprotocol.io) 接入外部工具（stdio + HTTP 传输），可从 Claude Desktop 导入
- **CLAUDE.md 记忆** — 自动加载项目目录树中所有 `CLAUDE.md` 文件作为上下文
- **会话管理** — 恢复历史对话；自动压缩过长的上下文历史
- **Vim 模式** — 输入框支持可选的 Vim 按键绑定
- **OAuth + API Key 认证** — `claude auth login` 走 Anthropic OAuth，或直接配置 `ANTHROPIC_API_KEY`
- **内置斜杠命令** — `/compact`、`/commit`、`/review`、`/model`、`/theme` 等（见下文）
- **流式响应** — 实时 token 流式输出，支持 thinking block 展示

## 架构设计

Claude Code Go 采用六层架构：

```
┌─────────────────────────────────────┐
│  CLI (cmd/claude)                   │  Cobra 入口
├─────────────────────────────────────┤
│  TUI (internal/tui)                 │  Bubble Tea MVU 界面
├─────────────────────────────────────┤
│  Tools (internal/tools)             │  文件、命令、搜索、MCP 工具
├─────────────────────────────────────┤
│  Core Engine (internal/engine)      │  流式推理、工具分发、多 Agent 协调
├─────────────────────────────────────┤
│  Services (internal/api, oauth,     │  Anthropic API、OAuth、MCP 客户端
│            mcp, compact)            │
├─────────────────────────────────────┤
│  Infra (pkg/types, internal/config, │  类型、配置、状态、钩子、插件
│         state, session, hooks)      │
└─────────────────────────────────────┘
```

详细说明见 [`docs/project/architecture.md`](docs/project/architecture.md)。

## 环境要求

- Go 1.21 或更高版本
- [Anthropic API Key](https://console.anthropic.com/) **或** Claude.ai 账号（OAuth）

## 安装

### 从源码编译

```bash
git clone https://github.com/808cn163/claude-code-go.git
cd claude-code-go
make build
# 产物路径：./bin/claude
```

将 `bin` 目录加入 `PATH`：

```bash
export PATH="$PATH:$(pwd)/bin"
```

### 使用 `go install`

```bash
go install github.com/808cn163/claude-code-go/cmd/claude@latest
```

## 快速开始

```bash
# 设置 API Key（或使用 OAuth，参见下方认证说明）
export ANTHROPIC_API_KEY=sk-ant-...

# 在当前目录启动交互式会话
claude

# 单次提问后退出
claude -p "解释一下这个项目的主入口"

# 恢复最近一次会话
claude --resume
```

## 认证

**API Key（推荐用于 CI/脚本）：**

```bash
export ANTHROPIC_API_KEY=sk-ant-...
# 或持久化保存：
claude auth login --api-key sk-ant-...
```

**OAuth（推荐用于交互式使用）：**

```bash
claude auth login    # 在浏览器中打开 OAuth 授权流程
claude auth status   # 查看当前登录身份
```

## API 提供商

Claude Code Go 支持多种 API 提供商，不仅可以使用 Anthropic 的 API，还可以使用 OpenAI 兼容的 API。

### 支持的提供商(`CLAUDE_PROVIDER=openai`)

| 提供商 | 说明 | 环境变量 |
|--------|------|----------|
| `direct`（默认） | Anthropic 直连 API | `ANTHROPIC_API_KEY`、`ANTHROPIC_BASE_URL` |
| `openai` | OpenAI 及兼容 API | `OPENAI_API_KEY`、`OPENAI_BASE_URL` |
| `bedrock` | AWS Bedrock | 通过环境变量设置 AWS 凭证 |
| `vertex` | Google Cloud Vertex AI | 通过环境变量设置 GCP 凭证 |

### 使用 Anthropic 兼容端点（中转/自建）

使用第三方中转或自建的 Anthropic 协议端点（需提供 `/v1/messages`）时，**不要**设置 `CLAUDE_PROVIDER=openai`——那会改走 OpenAI 协议分支。保持默认的 `direct` 即可：

```bash
export ANTHROPIC_BASE_URL=https://your-endpoint.com   # 末尾不要带斜杠
export ANTHROPIC_API_KEY=your-key
# export ANTHROPIC_MODEL=claude-sonnet-5              # 可选
```

- provider 取值为 `direct`（默认，走 Anthropic 协议）/`openai`/`bedrock`/`vertex`/`foundry`，没有 `anthropic`。
- 若只设置 `OPENAI_API_KEY` 而未设置 `ANTHROPIC_API_KEY`，provider 会被**自动改为 `openai`**，请显式设置 `ANTHROPIC_API_KEY`。
- 本项目仅以 `x-api-key` 头发送密钥，不支持 `Authorization: Bearer`；`ANTHROPIC_CUSTOM_HEADERS` 亦未实现。

### 使用 OpenAI 兼容 API

支持 OpenAI、DeepSeek、通义千问、Moonshot 或任何 OpenAI 兼容 API：

**方式一：环境变量**

```bash
# 设置提供商为 openai
export CLAUDE_PROVIDER=openai

# 设置 API Key
export OPENAI_API_KEY=sk-xxx

# 可选：设置自定义 Base URL（用于 OpenAI 兼容服务）
export OPENAI_BASE_URL=https://api.deepseek.com  # DeepSeek
# export OPENAI_BASE_URL=https://api.moonshot.cn/v1  # Moonshot
# export OPENAI_BASE_URL=http://localhost:11434/v1  # Ollama

# 设置模型
export OPENAI_MODEL=deepseek-chat

# 启动 Claude Code
claude
```

**方式二：配置文件**

创建或编辑 `~/.config/claude-code/settings.json`：

```json
{
  "provider": "openai",
  "apiKey": "sk-xxx",
  "baseUrl": "https://api.openai.com",
  "model": "gpt-4-turbo",
  "openaiOrganization": "org-xxx",
  "openaiProject": "proj-xxx"
}
```

### 各提供商配置示例

**OpenAI：**
- 支持所有 GPT-4 和 GPT-3.5 模型
- 完整的工具/函数调用支持
- 流式响应

**DeepSeek：**
```bash
export CLAUDE_PROVIDER=openai
export OPENAI_API_KEY=sk-xxx
export OPENAI_BASE_URL=https://api.deepseek.com
export OPENAI_MODEL=deepseek-chat
```

**Ollama（本地部署）：**
```bash
export CLAUDE_PROVIDER=openai
export OPENAI_BASE_URL=http://localhost:11434/v1
export OPENAI_MODEL=llama3
```

**Azure OpenAI：**
```bash
export CLAUDE_PROVIDER=openai
export OPENAI_API_KEY=your-azure-key
export OPENAI_BASE_URL=https://your-resource.openai.azure.com
export OPENAI_MODEL=your-deployment-name
```

**OrcaRouter（AI 网关）：**
[OrcaRouter](https://www.orcarouter.ai/ref/ref_49eee7ba2e9927450075) 是一个 AI 网关：把每个请求在 200+ 个模型（前沿与开源）中路由到最合适的一个，价格与各提供商官方价格一致，并提供故障转移与请求日志。它兼容 OpenAI API，直接通过 `openai` 提供商即可使用：

```bash
export CLAUDE_PROVIDER=openai
export OPENAI_API_KEY=<your-orcarouter-key>
export OPENAI_BASE_URL=https://api.orcarouter.ai/v1
export OPENAI_MODEL=orcarouter/auto   # 让 OrcaRouter 为每个请求挑选模型
```

## 使用说明

### 交互模式

```
claude [flags]
```

| 参数 | 说明 |
|------|------|
| `--resume` | 恢复最近一次会话 |
| `--session <id>` | 按 ID 恢复指定会话 |
| `--model <name>` | 覆盖默认 Claude 模型 |
| `--dark` / `--light` | 强制使用深色或浅色主题 |
| `--vim` | 启用 Vim 按键绑定 |
| `-p, --print <prompt>` | 非交互模式：执行单次提问后退出 |

### CLI 子命令

非交互式管理命令。`mcp` 与 `plugin` 两组的输出与 TypeScript 原版**字节级一致**：

| 命令 | 说明 |
|------|------|
| `claude mcp add <name> -- <command> [args…]` | 注册 stdio MCP 服务器（远程用 `--transport http <name> <url>`，环境变量用 `-e KEY=VAL`） |
| `claude mcp list` / `claude mcp get <name>` | 列出已配置服务器 / 健康检查单个服务器 |
| `claude mcp add-json <name> <json>` | 从原始 JSON 注册服务器 |
| `claude mcp add-from-claude-desktop` | 从 Claude Desktop 配置导入服务器 |
| `claude mcp remove <name>` / `reset-project-choices` | 移除服务器 / 重置项目级选择 |
| `claude plugin install <plugin>` | 从已配置的 marketplace 安装插件 |
| `claude plugin list` / `enable` / `disable` / `uninstall` / `update` | 管理已安装插件 |
| `claude plugin validate <path>` | 校验插件或 marketplace manifest（`--strict`、`--json`） |
| `claude plugin marketplace add/list/remove/update` | 管理插件 marketplace（本地路径） |
| `claude auth login` / `logout` / `status` | OAuth 或 API Key 认证 |
| `claude doctor` | 环境健康检查 |

### 斜杠命令

在输入框中输入 `/` 即可查看所有可用命令：

| 命令 | 说明 |
|------|------|
| `/help` | 显示所有命令 |
| `/clear` | 清空对话历史 |
| `/compact` | 压缩历史以减少上下文占用 |
| `/exit` | 退出 Claude Code |
| `/model` | 查看或切换当前模型 |
| `/theme` | 切换配色主题 |
| `/vim` | 切换 Vim 按键绑定 |
| `/effort` | 设置努力等级（low / medium / high） |
| `/status` | 查看会话状态 |
| `/cost` | 查看本次会话 token 用量 |
| `/session` | 查看当前会话 ID |
| `/memory` | 查看已加载的 CLAUDE.md 记忆文件 |
| `/dream` | 手动触发记忆整理 |
| `/review` | 让 Claude 评审当前 git diff |
| `/commit` | 让 Claude 撰写并创建 git commit |
| `/diff` | 查看当前 git diff |
| `/init` | 为本项目生成 CLAUDE.md |

`/config`、`/mcp`、`/resume`、`/terminal-setup` 已注册但尚未实现——MCP 管理请暂时使用上方的 CLI 子命令。

## 零人工代码的构建故事

> **本仓库中不存在任何人类编写的生产代码。**

整个项目由 **9 个 Claude AI Agent** 分工协作完成——从架构设计、详细设计文档、并行编码实现、代码评审，到 QA 验收与集成测试，全流程均由 AI 驱动：

```
PM Agent          →  项目计划、里程碑、任务调度
Tech Lead Agent   →  架构设计、设计文档评审、代码评审
Agent-Infra       →  基础设施层（类型、配置、状态、会话）
Agent-Services    →  服务层（API 客户端、OAuth、MCP、压缩）
Agent-Core        →  核心引擎（推理循环、工具分发、多 Agent 协调）
Agent-Tools       →  工具层（文件、命令、搜索、Web 等，构建期 18 个工具）
Agent-TUI         →  界面层（Bubble Tea MVU、主题、Vim 模式）
Agent-CLI         →  入口层（Cobra CLI、依赖注入、启动流程）
QA Agent          →  测试策略、逐层验收、集成测试
```

各 Agent 在独立 Git Worktree 分支上并行开发，通过共享代码库、设计文档和 QA 报告协作交互。最终在 **9 天内产出约 7,000 行生产代码 + 完整测试套件**，`go test -race ./...` 全部通过。

这是一次真实规模的验证：**非平凡的多层 Go 应用可以完全由 AI Agent 异步协作设计、实现、评审并交付**。2026 年 8 月起项目进入维护期，仍由 AI 主导——一个主 Claude 会话对照 TypeScript 原版的行为差距台账逐项关闭差距，人类负责人负责评审与合并。至今仍无任何人类编写的生产代码（当前约 33,000 行）。完整决策记录见 [`docs/project/`](docs/project/)。

## 开发指南

### 前置条件

- Go 1.21+
- `golangci-lint`（可选，用于代码检查）

### 构建与测试

```bash
# 构建
make build

# 运行所有测试
make test

# 生成覆盖率报告
make test-cover

# 静态检查
make vet

# 代码检查（需安装 golangci-lint）
make lint

# 构建 + 测试 + 静态检查
make all
```

### 项目结构

```
claude-code-go/
├── cmd/claude/          # CLI 入口
├── internal/
│   ├── api/             # Anthropic API 客户端与流式传输
│   ├── bootstrap/       # 应用初始化
│   ├── commands/        # 斜杠命令处理器
│   ├── compact/         # 对话压缩
│   ├── config/          # 配置（文件 + 环境变量）
│   ├── coordinator/     # 多 Agent 协调器
│   ├── engine/          # 推理引擎、工具分发
│   ├── hooks/           # 工具前/后钩子
│   ├── mcp/             # MCP 服务器管理
│   ├── memdir/          # CLAUDE.md 加载器
│   ├── oauth/           # OAuth2 流程
│   ├── permissions/     # 工具权限层
│   ├── plugin/          # 插件系统
│   ├── session/         # 会话持久化
│   ├── state/           # 应用状态
│   ├── tools/           # Tool 接口、注册表及内置工具实现
│   │   ├── agent/       #   子 Agent 与消息工具
│   │   ├── fileops/     #   文件读写/编辑/搜索工具
│   │   ├── interact/    #   用户交互与 Worktree 工具
│   │   ├── mcp/         #   MCP 工具适配器
│   │   ├── misc/        #   杂项工具
│   │   ├── shell/       #   Bash 执行工具
│   │   ├── tasks/       #   任务列表工具
│   │   └── web/         #   网页抓取与搜索工具
│   └── tui/             # Bubble Tea UI 组件
├── pkg/
│   └── types/           # 共享公共类型
├── docs/                # 设计文档与 QA 报告
├── Makefile
└── go.mod
```

## 路线图

初始构建（2026 年 4 月）在 9 天内交付了完整的六层架构。2026 年 8 月起项目进入**维护模式**：开发由行为对等差距台账（[`test/parity/cases.md`](test/parity/cases.md)）驱动——逐项对比本 CLI 与 TypeScript 原版的行为，每关闭一个差距都以字节级测试钉住，每一个有意为之的分歧都以 `waived` 记录原因。

近期里程碑：

- **2026-09-05** — `mcp`（7 个命令）与 `plugin`（8 个命令 + `marketplace` 子树）CLI 命令组落地，输出与 TS 原版字节级一致；用户可见桩 42 → 27
- **2026-08-30** — `--version` 格式对齐；对照 TS CLI 完成子命令全量盘点（§B11）
- **2026-08-29** — 维护模式流程落地：CI 强制的欠账红线、parity 测试骨架

原始的分阶段计划（v0.2.0 → v1.0.0）及各阶段完成状态保留在 [`docs/ROADMAP.md`](docs/ROADMAP.md)。

### 当前状态

```
✅ 已完成: 核心引擎 + 流式输出、TUI、34 个工具、17 个斜杠命令（另有 4 个桩）、
          mcp/plugin/auth/doctor CLI 子命令、9 层权限流水线、
          三级上下文压缩（snip/micro/auto）、OAuth、会话持久化
🔧 模式:  维护期——以差距台账驱动的对等修复（对照 TypeScript CLI）
📉 欠账:  56 个 TODO(dep) / 27 个用户可见桩，CI 红线只降不升
🚀 发布:  v0.8.0，5 平台二进制
```

## 贡献指南

欢迎贡献！提交 Pull Request 前请先阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。

快速检查清单：
- Fork 仓库并创建功能分支
- 确保 `make test` 和 `make vet` 通过
- 为新功能编写测试
- 遵循现有代码风格（运行 `gofmt ./...`）
- 使用提供的模板提交 Pull Request

## 安全

如需报告安全漏洞，请参阅 [SECURITY.md](SECURITY.md)。**请勿**在公开的 GitHub Issue 中披露安全问题。

## 相对上游 Claude Code 的限制调整

本项目对官方 Claude Code 的限制移除做法，提供了对应的**机制层放宽开关**。

### 对照 ClawGod 审计

ClawGod 1.8.0 从官方 Claude Code 中移除 4 项限制。逐条比对本项目后的结论：

| ClawGod 移除项 | 本项目情况 | 说明 |
|---|---|---|
| CYBER_RISK（安全测试拒绝策略） | 无提示词层对应，有机制层对应 | 本项目**没有 base system prompt**（交互模式的 system prompt 仅由 CLAUDE.md 构成），故不存在该提示词。机制层对应物是危险命令拦截表 |
| URL 生成限制 | **不存在** | 全仓库无 `You must NEVER generate or guess URLs` 类文本 |
| Cautious Actions（操作前强制确认） | 无提示词层对应，有机制层对应 | 机制层对应物是权限管线的 `Ask` 决策 |
| Not logged in 横幅 | **无等价物** | 仅有 `internal/bootstrap/auth.go` 的认证失败错误提示，性质不同（见下方「未改动项」） |

> 换言之：本项目以**代码机制*的形式实现了约束。因此本项目的对应处理是放宽机制，而非删除文本。

### 放宽开关

| 项 | 内容 |
|---|---|
| 开关名 | `CLAUDE_CODE_RELAXED_LIMITS` |
| 取值 | `1` / `true` / `yes` / `on`（由 `pkg/utils/env.IsEnvTruthy` 判定） |
| 定义位置 | `internal/config/relaxed.go` → `RelaxedLimitsEnabled()`（全项目唯一判定入口） |
| 默认值 | **关闭**。未设置时所有行为与改动前**完全一致**，现有安全姿态不变 |

启用方式：

```bash
export CLAUDE_CODE_RELAXED_LIMITS=1
claude
```

### 开关开启后的变化

| 文件 | 改动点 | 效果 |
|------|--------|------|
| `internal/tools/shell/security.go` | `AnalyzeCommand()` 入口条件提前返回 `nil` | `sudo`、`rm -rf`、`mkfs`、`dd of=/dev/`、`curl \| sh` 等不再被拦截 |
| `internal/permissions/checker.go` | `CanUseTool()` 单点把 `Ask` 降级为 `Allow` | 破坏性操作不再弹出确认提示 |

**关键边界**：只降级 `Ask`，**从不**降级 `Deny`。显式 deny 规则、validate/hook 拒绝、plan 模式的写入拦截、工具自身的拒绝，在开关开启时依然生效——该开关**无法**用来绕过一次明确的拒绝。

实现上的两点克制：

- 降级集中在 `CanUseTool()` **唯一一处**，而不是散落在产生 `Ask` 的各分支里，避免后续新增分支时遗漏。
- `dangerousCommandPatterns` 拦截表**未被修改**，仅在其入口增加条件提前返回；关闭开关即完整恢复原行为。

### 为什么默认关闭

1. 放宽安全限制应是**显式选择**，而非升级后的默认后果。
2. 本项目的默认行为有 CI 红线约束（`make test`、`make debt-check`），默认放宽会使既有测试与既有安全契约失效。
3. 开关化后，同一份代码可按环境切换姿态，无需维护分支。

## 许可证

本项目基于 MIT 许可证——详见 [LICENSE](LICENSE)。

## 相关项目

- [claude-code](https://github.com/anthropics/claude-code) — 原版 TypeScript CLI
- [anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) — Anthropic API 官方 Go SDK
- [Model Context Protocol](https://modelcontextprotocol.io) — AI 接入工具的开放标准
