package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// appVersion is the version string for the welcome banner.
// This should be kept in sync with bootstrap.appVersion.
const appVersion = "0.1.0"

// productName 是产品中文名。
const productName = "Claude Code 网络安全专用版"

// productVersion 是产品展示版本号。
const productVersion = "1.0"

// WelcomeHeader contains data for the startup welcome banner.
type WelcomeHeader struct {
	version string
	model   string
	cwd     string
	shown   bool // whether the header has been shown (only show once)
}

// NewWelcomeHeader creates a new welcome header with default values.
func NewWelcomeHeader(model, cwd string) WelcomeHeader {
	return WelcomeHeader{
		version: appVersion,
		model:   model,
		cwd:     cwd,
		shown:   false,
	}
}

// MarkShown marks the header as shown.
func (w WelcomeHeader) MarkShown() WelcomeHeader {
	w.shown = true
	return w
}

// IsShown returns whether the header has been shown.
func (w WelcomeHeader) IsShown() bool {
	return w.shown
}

// View renders the welcome header banner.
// Format:
//
//	Claude Code 网络安全专用版 V1.0
//	claude-sonnet-4-20250514 · API 计费
//	~/path/to/cwd
//	欢迎使用 Claude Code 网络安全专用版！  /effort 调节速度与智能的权衡
func (w WelcomeHeader) View(width int, theme Theme) string {
	if w.shown {
		return ""
	}

	var sb strings.Builder

	// Claude Code 官方图标（ASCII 艺术字）
	logo := renderLogo(theme)

	// 版本行（产品中文名 + 展示版本号）
	versionLine := primaryStyle(theme).Bold(true).Render(
		fmt.Sprintf("%s V%s", productName, productVersion))

	// Model + billing info
	modelStr := w.model
	if modelStr == "" {
		modelStr = "claude-sonnet-4-20250514"
	}
	modelLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		secondaryStyle(theme).Render(modelStr),
		mutedStyle(theme).Render(" · "),
		mutedStyle(theme).Render("API 计费"),
	)

	// Working directory
	cwdLine := mutedStyle(theme).Render(shortenPath(w.cwd))

	// Welcome message
	welcomeLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		successStyle(theme).Render("欢迎使用 Claude Code 网络安全专用版！"),
		mutedStyle(theme).Render("  "),
		accentStyle(theme).Render("/effort"),
		mutedStyle(theme).Render(" 调节速度与智能的权衡"),
	)

	// 标题 / 模型 / 工作目录三行与 logo 顶部对齐（logo 恰好 3 行）
	infoBlock := lipgloss.JoinVertical(
		lipgloss.Left,
		versionLine,
		modelLine,
		cwdLine,
	)

	// 欢迎语单独置于横幅下方：若并入 infoBlock 会使其变成 4 行，
	// 导致 3 行的 logo 被居中后与标题行错开。
	banner := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, logo, "  ", infoBlock),
		welcomeLine,
	)

	sb.WriteString(banner)
	sb.WriteString("\n")

	return sb.String()
}

// renderLogo 渲染 Claude Code 官方图标的 ASCII 艺术字。
// 恒为天蓝色，不随主题变化。注意：每一行开头的空格用于图案左右对齐，必须原样保留。
func renderLogo(_ Theme) string {
	// Claude Code 官方图标，精确 3 行
	logoLines := []string{
		" ▐▛███▜▌",
		"▝▜█████▛▘",
		"  ▘▘ ▝▝",
	}

	// 固定为蓝色，直接输出 ANSI 转义序列而非走 lipgloss 样式：lipgloss 会依据
	// termenv 探测到的终端色阶做降级，而该探测在 WSL / SSH / 容器环境下常把终端
	// 误判为低色阶，把蓝色降级成亮青（深色背景上近似白色）。
	// 34 为标准蓝、94 为亮蓝，按需替换即可。
	const logoColor = "\x1b[94m"
	const logoReset = "\x1b[0m"

	rendered := make([]string, 0, len(logoLines))
	for _, line := range logoLines {
		rendered = append(rendered, logoColor+line+logoReset)
	}
	return strings.Join(rendered, "\n")
}

// shortenPath shortens a path for display, replacing home dir with ~.
func shortenPath(path string) string {
	// This is a simple implementation - could be enhanced to use os.UserHomeDir()
	if strings.HasPrefix(path, "/Users/") {
		parts := strings.SplitN(path, "/", 4)
		if len(parts) >= 4 {
			return "~/" + parts[3]
		}
	}
	if strings.HasPrefix(path, "/home/") {
		parts := strings.SplitN(path, "/", 4)
		if len(parts) >= 4 {
			return "~/" + parts[3]
		}
	}
	return path
}
