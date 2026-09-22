// Package api provides the OpenAI-compatible API client implementation.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// openaiClient implements Client using the OpenAI Chat Completions API format.
// This enables compatibility with OpenAI, Azure OpenAI, and other OpenAI-compatible services
// like DeepSeek, Qwen, Moonshot, etc.
type openaiClient struct {
	apiKey       string
	baseURL      string
	httpClient   *http.Client
	headers      map[string]string
	organization string       // Optional OpenAI organization ID
	project      string       // Optional OpenAI project ID
	debugLogger  *DebugLogger // Debug logger for request/response tracing
}

// Compile-time interface assertion: openaiClient must satisfy Client.
var _ Client = (*openaiClient)(nil)

// newOpenAIClient creates a new OpenAI-compatible client.
func newOpenAIClient(cfg ClientConfig, httpClient *http.Client, headers map[string]string) (Client, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	dl, err := NewDebugLogger(cfg.Debug, cfg.DebugFile)
	if err != nil {
		return nil, fmt.Errorf("api: init debug logger: %w", err)
	}
	return &openaiClient{
		apiKey:       cfg.APIKey,
		baseURL:      baseURL,
		httpClient:   httpClient,
		headers:      headers,
		organization: cfg.OpenAIOrganization,
		project:      cfg.OpenAIProject,
		debugLogger:  dl,
	}, nil
}

// Stream initiates a streaming chat completion request and returns an SSE event stream.
func (c *openaiClient) Stream(ctx context.Context, req *MessageRequest) (StreamReader, error) {
	req.Stream = true
	httpReq, err := c.buildRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.debugLogger.LogError("HTTP request failed", err)
		return nil, &APIError{Kind: ErrKindConnectionError, Message: err.Error()}
	}
	// Log response status and headers.
	c.debugLogger.LogResponse(resp.StatusCode, resp.Status, resp.Header)
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		apiErr := c.parseHTTPError(resp)
		c.debugLogger.LogError("API returned error", apiErr)
		return nil, apiErr
	}
	return newOpenAISSEReader(resp, c.debugLogger), nil
}

// Complete makes a non-streaming chat completion request and returns the full response.
func (c *openaiClient) Complete(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	req.Stream = false
	httpReq, err := c.buildRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.debugLogger.LogError("HTTP request failed", err)
		return nil, &APIError{Kind: ErrKindConnectionError, Message: err.Error()}
	}
	defer resp.Body.Close()
	// Log response status and headers.
	c.debugLogger.LogResponse(resp.StatusCode, resp.Status, resp.Header)
	if resp.StatusCode >= 400 {
		apiErr := c.parseHTTPError(resp)
		c.debugLogger.LogError("API returned error", apiErr)
		return nil, apiErr
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.debugLogger.LogError("read response body", err)
		return nil, fmt.Errorf("api: read openai response body: %w", err)
	}
	c.debugLogger.LogResponseBody(respBody)
	var openaiResp openaiChatCompletionResponse
	if decErr := json.Unmarshal(respBody, &openaiResp); decErr != nil {
		c.debugLogger.LogError("decode response JSON", decErr)
		return nil, fmt.Errorf("api: decode openai response: %w", decErr)
	}
	return c.convertToMessageResponse(&openaiResp), nil
}

// buildRequest constructs the HTTP request for OpenAI API.
func (c *openaiClient) buildRequest(ctx context.Context, req *MessageRequest) (*http.Request, error) {
	openaiReq := c.convertToOpenAIRequest(req)
	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("api: marshal openai request: %w", err)
	}
	url := c.baseURL + "/v1/chat/completions"

	// Build headers map for logging.
	reqHeaders := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + c.apiKey,
	}
	if c.organization != "" {
		reqHeaders["OpenAI-Organization"] = c.organization
	}
	if c.project != "" {
		reqHeaders["OpenAI-Project"] = c.project
	}
	for k, v := range c.headers {
		reqHeaders[k] = v
	}

	// Log the full request.
	c.debugLogger.LogRequest(http.MethodPost, url, reqHeaders, body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("api: build openai request: %w", err)
	}
	for k, v := range reqHeaders {
		httpReq.Header.Set(k, v)
	}
	return httpReq, nil
}

// convertToOpenAIRequest converts Anthropic MessageRequest to OpenAI ChatCompletionRequest.
func (c *openaiClient) convertToOpenAIRequest(req *MessageRequest) *openaiChatCompletionRequest {
	openaiReq := &openaiChatCompletionRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Stream:    req.Stream,
	}

	// Convert system message
	if req.System != "" {
		sysContent := req.System
		openaiReq.Messages = append(openaiReq.Messages, openaiMessage{
			Role:    "system",
			Content: &sysContent,
		})
	}

	// Convert messages
	for _, msg := range req.Messages {
		openaiMsg := c.convertMessage(msg)
		openaiReq.Messages = append(openaiReq.Messages, openaiMsg...)
	}

	// Convert tools
	if len(req.Tools) > 0 {
		openaiReq.Tools = make([]openaiTool, 0, len(req.Tools))
		for _, tool := range req.Tools {
			openaiReq.Tools = append(openaiReq.Tools, openaiTool{
				Type: "function",
				Function: openaiFunction{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			})
		}
	}

	return openaiReq
}

// convertMessage converts an Anthropic message to OpenAI message(s).
func (c *openaiClient) convertMessage(msg MessageParam) []openaiMessage {
	var result []openaiMessage

	// Try to unmarshal content as string first
	var textContent string
	if err := json.Unmarshal(msg.Content, &textContent); err == nil {
		result = append(result, openaiMessage{
			Role:    c.convertRole(msg.Role),
			Content: &textContent,
		})
		return result
	}

	// Try to unmarshal as array of content blocks
	var blocks []ContentBlock
	if err := json.Unmarshal(msg.Content, &blocks); err != nil {
		// Fallback: treat as raw string
		fallback := string(msg.Content)
		result = append(result, openaiMessage{
			Role:    c.convertRole(msg.Role),
			Content: &fallback,
		})
		return result
	}

	// For assistant messages with both text and tool_use, combine them
	if msg.Role == "assistant" {
		var combinedText string
		var toolCalls []openaiToolCall

		for _, block := range blocks {
			switch block.Type {
			case "text":
				if combinedText != "" {
					combinedText += "\n"
				}
				combinedText += block.Text
			case "tool_use":
				// block.Input 本身已是 JSON 片段。若模型曾产出非法 JSON，
				// 直接透传会让后续整体序列化失败，故校验后回退为空对象。
				args := block.Input
				if len(args) == 0 || !json.Valid(args) {
					args = json.RawMessage("{}")
				}
				// 注意：请求体的 tool_calls 不带 index 字段
				// （index 只属于流式响应的 delta）。
				toolCalls = append(toolCalls, openaiToolCall{
					ID:   block.ID,
					Type: "function",
					Function: openaiToolCallFunction{
						Name:      block.Name,
						Arguments: string(args),
					},
				})
			case "thinking":
				// Anthropic 的 thinking 在 Chat Completions 协议中没有对应概念。
				// 这里刻意丢弃：若回传，思考内容会被当作助手正文，且会生成连续的
				// assistant 消息，违反 OpenAI 的消息交替约束。
			}
		}

		// Create assistant message with combined content and tool calls
		if combinedText != "" || len(toolCalls) > 0 {
			// Always set Content for assistant messages with tool_calls.
			// Many OpenAI-compatible APIs require the content field to be present
			// (even if empty) when tool_calls are included.
			contentPtr := &combinedText
			assistantMsg := openaiMessage{
				Role:    "assistant",
				Content: contentPtr,
			}
			if len(toolCalls) > 0 {
				assistantMsg.ToolCalls = toolCalls
			}
			result = append(result, assistantMsg)
		}

		return result
	}

	// 非 assistant 消息按协议要求分两趟输出：
	// 同一条消息内的 tool_result 必须先于文本，否则 role=tool 会被 role=user 隔断，
	// 破坏「role=tool 必须紧跟 assistant(tool_calls)」的约束，严格实现会直接报错。
	// 第一趟：tool_result。
	for _, block := range blocks {
		if block.Type != "tool_result" {
			continue
		}
		// Content 可以是字符串，也可以是内容块数组。
		contentStr := extractToolResultContent(block.Content)
		if c.debugLogger.Enabled() {
			preview := contentStr
			if len(preview) > 200 {
				preview = preview[:200]
			}
			c.debugLogger.Logf("tool_result: ToolUseID=%q, ContentLen=%d, Preview=%q", block.ToolUseID, len(contentStr), preview)
		}
		result = append(result, openaiMessage{
			Role:       "tool",
			Content:    &contentStr,
			ToolCallID: block.ToolUseID,
		})
	}

	// 第二趟：文本（跳过空块，避免产生 content 为空的消息）。
	for _, block := range blocks {
		if block.Type != "text" || block.Text == "" {
			continue
		}
		text := block.Text
		result = append(result, openaiMessage{
			Role:    c.convertRole(msg.Role),
			Content: &text,
		})
	}

	return result
}

// extractToolResultContent extracts text content from tool_result.
// Content can be:
// - A simple string: "result text"
// - An array of content blocks: [{"type":"text","text":"result text"}]
func extractToolResultContent(content json.RawMessage) string {
	if len(content) == 0 {
		return ""
	}

	// Try to unmarshal as string first
	var strContent string
	if err := json.Unmarshal(content, &strContent); err == nil {
		return strContent
	}

	// Try to unmarshal as array of content blocks
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(content, &blocks); err == nil {
		var texts []string
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				texts = append(texts, b.Text)
			}
		}
		if len(texts) > 0 {
			return joinStrings(texts, "\n")
		}
	}

	// Fallback: return raw JSON string
	return string(content)
}

// joinStrings joins non-empty strings with separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// convertRole converts Anthropic role to OpenAI role.
func (c *openaiClient) convertRole(role string) string {
	switch role {
	case "user":
		return "user"
	case "assistant":
		return "assistant"
	default:
		return role
	}
}

// convertToMessageResponse converts OpenAI response to Anthropic MessageResponse.
func (c *openaiClient) convertToMessageResponse(resp *openaiChatCompletionResponse) *MessageResponse {
	msgResp := &MessageResponse{
		ID:    resp.ID,
		Type:  "message",
		Role:  "assistant",
		Model: resp.Model,
		Usage: Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}

	// Convert stop reason
	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		switch choice.FinishReason {
		case "stop":
			msgResp.StopReason = "end_turn"
		case "length":
			msgResp.StopReason = "max_tokens"
		case "tool_calls":
			msgResp.StopReason = "tool_use"
		default:
			msgResp.StopReason = choice.FinishReason
		}

		// Convert content
		if choice.Message.Content != "" {
			msgResp.Content = append(msgResp.Content, ContentBlock{
				Type: "text",
				Text: choice.Message.Content,
			})
		}

		// Convert tool calls
		for _, tc := range choice.Message.ToolCalls {
			// arguments 是 JSON 字符串，需转成对象的 JSON 片段。
			// 模型偶发非法 JSON 时不能直接透传：json.RawMessage 要求内容合法，
			// 否则后续整体序列化会失败，污染整个会话历史。
			input := json.RawMessage("{}")
			if tc.Function.Arguments != "" {
				if json.Valid([]byte(tc.Function.Arguments)) {
					input = json.RawMessage(tc.Function.Arguments)
				} else {
					c.debugLogger.Logf("tool_call %q: arguments 不是合法 JSON，已回退为空对象: %q",
						tc.ID, tc.Function.Arguments)
				}
			}
			msgResp.Content = append(msgResp.Content, ContentBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: input,
			})
		}
	}

	return msgResp
}

// parseHTTPError reads an error response body and returns an *APIError.
func (c *openaiClient) parseHTTPError(resp *http.Response) *APIError {
	body, _ := io.ReadAll(resp.Body)
	var errBody struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	msg := string(body)
	if json.Unmarshal(body, &errBody) == nil && errBody.Error.Message != "" {
		msg = errBody.Error.Message
	}
	kind := classifyOpenAIError(resp.StatusCode, msg)
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    msg,
		Headers:    resp.Header,
		Kind:       kind,
	}
}

// classifyOpenAIError maps OpenAI HTTP status codes to APIErrorKind.
func classifyOpenAIError(statusCode int, message string) APIErrorKind {
	switch statusCode {
	case 401:
		return ErrKindUnauthorized
	case 403:
		return ErrKindForbidden
	case 429:
		return ErrKindRateLimit
	case 400:
		if contextOverflowRe.MatchString(message) {
			return ErrKindContextWindow
		}
		return ErrKindInvalidRequest
	default:
		if statusCode >= 500 {
			return ErrKindServerError
		}
		return ErrKindUnknown
	}
}
