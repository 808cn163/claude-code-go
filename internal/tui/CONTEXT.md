---
package: tui
import_path: internal/tui
layer: tui
generated_at: 2026-09-05T09:11:12Z
source_files: [agentcolors.go, cmds.go, coordinator.go, init.go, input.go, keys.go, messagelist.go, messages.go, model.go, permissions.go, spinner.go, statusbar.go, styles.go, theme.go, update.go, view.go, welcome.go]
---

# internal/tui

> Layer: **TUI** · Files: 17 · Interfaces: 0 · Structs: 35 · Functions: 6

## Structs

- **AgentColorManager** — 3 fields
- **AgentProgressMsg** — 3 fields: TaskID, Activity, Detail
- **AgentStatusMsg** — 3 fields: TaskID, Status, Description
- **AgentTaskState** — 12 fields: ID, Name, AgentType, Description, Status, StartTime, ElapsedMs, OutputTokens, ...
- **AppModel** — 45 fields
- **CommandResultMsg** — 2 fields: Text, IsError
- **CompactDoneMsg** — 1 fields: Summary
- **CoordinatorPanel** — 3 fields: Tasks, SelectedIndex, TaskOrder
- **DreamDoneMsg** — 1 fields: Err
- **InputChangedMsg** — 1 fields: Text
- **InputModel** — 4 fields
- **InputSubmittedMsg** — 1 fields: Text
- **MemdirLoadedMsg** — 2 fields: Paths, ScopedFiles
- **MessageLookups** — 4 fields: ToolUseToResult, ResolvedToolUseIDs, ErroredToolUseIDs, InProgressToolUseIDs
- **PermissionDialog** — 8 fields
- **PermissionRequestMsg** — 7 fields: RequestID, ToolName, ToolUseID, Message, Input, ProjectPath, RespFn
- **SlashCommandMsg** — 2 fields: Name, Args
- **SpinnerModel** — 4 fields
- **StatusBar** — 6 fields
- **StreamAssistantTurnMsg** — 1 fields: FinalMessage
- **StreamDoneMsg** — 1 fields: FinalMessage
- **StreamErrorMsg** — 1 fields: Err
- **StreamThinkingMsg** — 1 fields: Delta
- **StreamTokenMsg** — 1 fields: Delta
- **StreamToolResultMsg** — 3 fields: ToolUseID, Content, IsError
- **StreamToolUseCompleteMsg** — 2 fields: ToolUseID, ToolInput
- **StreamToolUseInputDeltaMsg** — 3 fields: ToolUseID, ToolName, InputDelta
- **StreamToolUseStartMsg** — 3 fields: ToolUseID, ToolName, InputDelta
- **StreamUserTurnMsg** — 1 fields: FinalMessage
- **SystemTextMsg** — 1 fields: Text
- **TermResizedMsg** — 2 fields: Width, Height
- **Theme** — 11 fields: Primary, Secondary, Accent, Muted, Error, Warning, Success, CodeBG, ...
- **TickMsg** — 1 fields: Time
- **TokenUsage** — 4 fields: Input, Output, CacheRead, CacheCreated
- **WelcomeHeader** — 4 fields

## Functions

- `IsSlashCommand(text string) bool`
- `MessageListView(messages []types.Message, width int, darkMode bool, theme Theme, mdRenderer *glamour.TermRenderer, expandedToolResults map[string]bool) string`
- `New(qe engine.QueryEngine, appStore *state.AppStateStore, vimEnabled bool, dark bool, permAskCh <-chan permissions.AskRequest, permRespCh chan<- permissions.AskResponse, agentCoord tools.AgentCoordinator, agentEventCh <-chan coordinator.Event, mq *msgqueue.MessageQueue, qg *msgqueue.QueryGuard, memoryStore *memdir.MemoryStore) tea.Model`
- `NewAgentColorManager() *AgentColorManager`
- `NewInput(vimEnabled bool) InputModel`
- `NewWelcomeHeader(model string, cwd string) WelcomeHeader`

## Constants

- `AgentCompleted`
- `AgentFailed`
- `AgentPaused`
- `AgentRunning`
- `PermissionChoiceAlwaysAllow`
- `PermissionChoiceNo`
- `PermissionChoiceYes`
- `SpinnerModeBrief`
- `SpinnerModeNormal`
- `SpinnerModeTeammate`
- `ToolStatusError`
- `ToolStatusInProgress`
- `ToolStatusQueued`
- `ToolStatusResolved`
- `VimModeInsert`
- `VimModeNormal`
- `VimModeVisual`

## Dependencies

**Imports:** `internal/commands`, `internal/coordinator`, `internal/engine`, `internal/memdir`, `internal/msgqueue`, `internal/permissions`, `internal/state`, `internal/tools`, `pkg/types`

**Imported by:** `internal/bootstrap`

<!-- AUTO-GENERATED ABOVE — DO NOT EDIT -->
<!-- MANUAL NOTES BELOW — preserved across regeneration -->

## Design Notes

- **输入框占位符必须以 ASCII 字符开头**（当前值 `[输入消息…]`）。bubbles v0.20.0 的
  `textarea.placeholderView()` 用 `plines[0][0]`（**字节索引**）取占位符首字符，再做
  `string(byte)` 转换；若首字符是多字节 UTF-8（如中文「输」E8 BE 93），会被拆成乱码
  （`è` + 0xBE + U+0093）。首字符为 ASCII 即可完全规避该上游缺陷，其后的中文均正常。
  用户实际**输入**的中文不受影响（只有占位符走这条渲染路径）。升级 bubbles 后请重测。
- **刻意不在 `bootstrap/root.go` 启用 `tea.WithMouseCellMotion()`**：启用鼠标追踪会让终端
  把鼠标事件全部转发给程序，用户就无法用鼠标拖选界面文字进行复制（终端原生选择失效）。
  滚动改由 PgUp/PgDn 承担；`keys.go` 的 `handleMouse` 保留以便将来需要时可恢复。
- **欢迎横幅 logo 与信息块顶部对齐**（`JoinHorizontal(lipgloss.Top, ...)`）：logo 为 3 行，
  若把欢迎语并入 `infoBlock` 会使其变成 4 行，`JoinHorizontal(lipgloss.Center)` 会把 logo
  挤得比标题行低半行，故欢迎语单独置于横幅下方。
- **logo 的蓝色是手写 ANSI 序列，不是笔误**：`renderLogo` 刻意不走 lipgloss 样式，而是
  直接拼 `"\x1b[94m"`（亮蓝；换成 `34` 即标准蓝）。原因是 lipgloss 会依据 termenv 探测到的
  终端色阶做降级，而该探测在 WSL / SSH / 容器环境下常把终端误判为低色阶，把蓝色降级成
  亮青（bright cyan，深色背景上肉眼近似白色）。手写序列可绕过这层降级，直接把色码交给终端。
