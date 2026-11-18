// 主系统提示词，与 claude-cli (claude 2.1.110) 抓包对齐。
// 模板中的 {{占位符}} 由 RenderClaudeCliSystemPrompt 填充。
package engine

import (
	_ "embed"
	"os"
	"runtime"
	"strings"
)

//go:embed prompt_template.txt
var claudeCliPromptTemplate string

// ccVersion / ccEntrypoint 镜像 claude-cli 的 attribution 头。
const (
	ccVersion   = "2.1.110"
	ccEntrypoint = "sdk-cli"
)

// RenderClaudeCliSystemPrompt 渲染与 claude-cli 一致的 system 块序列：
//
//	[0] x-anthropic-billing-header（无缓存标记）
//	[1] SDK 声明（无缓存标记）
//	[2] 主提示词（cache_control: ephemeral）
//	[3] Environment 段（动态，无缓存标记）——已并入模板尾部，此处返回 3 块。
func RenderClaudeCliSystemPrompt(cwd, model string) []SystemPromptPart {
	body := claudeCliPromptTemplate

	// Environment 动态段占位符。
	body = strings.ReplaceAll(body, "{{CWD}}", cwd)
	body = strings.ReplaceAll(body, "{{IS_GIT_REPO}}", boolWord(isGitRepo(cwd)))
	body = strings.ReplaceAll(body, "{{PLATFORM}}", runtime.GOOS)
	body = strings.ReplaceAll(body, "{{SHELL}}", shellName())
	body = strings.ReplaceAll(body, "{{OS_VERSION}}", osVersion())
	body = strings.ReplaceAll(body, "{{MODEL}}", model)
	body = strings.ReplaceAll(body, "{{MEMORY_DIR}}", memoryDirFor(cwd))

	parts := []SystemPromptPart{
		{Text: "x-anthropic-billing-header: cc_version=" + ccVersion + "." + fingerprint(cwd) +
			"; cc_entrypoint=" + ccEntrypoint + "; cch=00000;"},
		{Text: "You are a Claude agent, built on Anthropic's Claude Agent SDK."},
		{Text: body, CacheControl: "ephemeral"},
	}
	return parts
}

// boolWord returns "true"/"false" for b.
func boolWord(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// isGitRepo reports whether dir (or an ancestor) contains a .git directory.
func isGitRepo(dir string) bool {
	for d := dir; ; {
		if _, err := os.Stat(d + "/.git"); err == nil {
			return true
		}
		parent := d[:strings.LastIndexByte(d, '/')]
		if parent == d || parent == "" {
			return false
		}
		d = parent
	}
}

// shellName returns $SHELL 的 basename。
func shellName() string {
	sh := os.Getenv("SHELL")
	if sh == "" {
		return "bash"
	}
	if i := strings.LastIndexByte(sh, '/'); i >= 0 {
		return sh[i+1:]
	}
	return sh
}

// osVersion returns "Linux <kernel release>"（uname -srm 风格，对齐 claude-cli）。
func osVersion() string {
	if v, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		return "Linux " + strings.TrimSpace(string(v))
	}
	return "Linux " + runtime.GOOS
}

// memoryDirFor mirrors claude-cli 的项目记忆目录推导：
// ~/.claude/projects/<cwd 去斜杠连字符>/memory/
func memoryDirFor(cwd string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return cwd + "/.claude/memory/"
	}
	slug := strings.ReplaceAll(strings.Trim(cwd, "/"), "/", "-")
	return home + "/.claude/projects/-" + slug + "/memory/"
}

// fingerprint 镜像 claude-cli 的 cc_version 指纹（会话内稳定的短哈希）。
// claude-cli 用 msg 内容+版本计算；这里用 cwd+版本派生一个等长的确定性值。
func fingerprint(cwd string) string {
	sum := 0
	for i := 0; i < len(cwd); i++ {
		sum = sum*31 + int(cwd[i])
	}
	const hexDigits = "0123456789abcdef"
	out := make([]byte, 0, 3)
	for i := 0; i < 3; i++ {
		out = append(out, hexDigits[(sum>>(uint(i)*4))&0xf])
	}
	return string(out)
}
