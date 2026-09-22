package types

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestContentBlockToolUseMarshalAlwaysHasInput 确保 tool_use 块始终序列化出 input 字段，
// 且其为 JSON 对象。
//
// 严格的上游网关（如第三方 ANTHROPIC_BASE_URL 中转，serde 反序列化）会因该字段缺失
// 返回 422 "missing field `input`"；Messages 协议亦要求 input 是对象而非 null。
func TestContentBlockToolUseMarshalAlwaysHasInput(t *testing.T) {
	id, name := "toolu_1", "Bash"

	tests := []struct {
		name  string
		input map[string]any
	}{
		{"nil input", nil},
		{"empty map input", map[string]any{}},
		{"non-empty input", map[string]any{"command": "ls"}},
	}

	for _, tc := range tests {
		tc := tc // Go 1.21 语义：并行子测试需显式捕获循环变量
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			blk := ContentBlock{Type: ContentTypeToolUse, ID: &id, Name: &name, Input: tc.input}
			raw, err := json.Marshal(blk)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var decoded map[string]any
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			input, ok := decoded["input"]
			if !ok {
				t.Fatalf("input field missing in %s", raw)
			}
			if _, ok := input.(map[string]any); !ok {
				t.Errorf("input should be a JSON object, got %T in %s", input, raw)
			}
		})
	}
}

// TestContentBlockNonToolUseUnaffected 确保 MarshalJSON 不影响其它类型的块：
// text 块不应凭空多出 input 字段，且原有字段保持不变。
func TestContentBlockNonToolUseUnaffected(t *testing.T) {
	t.Parallel()

	text := "hello"
	raw, err := json.Marshal(ContentBlock{Type: ContentTypeText, Text: &text})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if strings.Contains(string(raw), "input") {
		t.Errorf("text block should not contain input field: %s", raw)
	}
	if !strings.Contains(string(raw), `"text":"hello"`) {
		t.Errorf("text block should keep its text field: %s", raw)
	}
}
