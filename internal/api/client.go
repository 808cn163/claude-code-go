// Package api provides the Anthropic API client implementation.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is the unified Anthropic API entry point.
// Supports Direct, Bedrock, Vertex, and Foundry providers.
type Client interface {
	// Stream initiates a streaming message request and returns an SSE event stream.
	// The caller is responsible for closing the StreamReader.
	Stream(ctx context.Context, req *MessageRequest) (StreamReader, error)

	// Complete makes a non-streaming (synchronous) request and returns the full response.
	Complete(ctx context.Context, req *MessageRequest) (*MessageResponse, error)
}

// StreamReader wraps SSE event stream reading, following io.Closer semantics.
type StreamReader interface {
	// Next returns the next SSE event; returns (nil, io.EOF) when the stream ends.
	Next() (*StreamEvent, error)
	// Close closes the underlying connection and releases resources. Idempotent.
	io.Closer
}

// MessageParam is a single message in the conversation.
type MessageParam struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"` // string or []ContentBlock
}

// ContentBlock is a single content block (text, tool_use, tool_result, image, thinking).
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
	// thinking
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// ToolSchema describes a tool for the API request.
type ToolSchema struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// SystemBlock is one block of the system prompt (mirrors the Anthropic API
// system array form). CacheControl mirrors cache_control on the last block.
type SystemBlock struct {
	Type         string `json:"type"`
	Text         string `json:"text"`
	CacheControl string `json:"cache_control,omitempty"`
}

// RequestMetadata mirrors the Anthropic API metadata field (user_id).
type RequestMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

// ThinkingConfig mirrors the Anthropic API thinking field
// ({type:"adaptive"} or {type:"enabled",budget_tokens:N}).
type ThinkingConfig struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

// MessageRequest corresponds to the Anthropic Messages API request body.
type MessageRequest struct {
	Model          string         `json:"model"`
	MaxTokens      int            `json:"max_tokens"`
	Messages       []MessageParam `json:"messages"`
	System         string         `json:"system,omitempty"`
	// SystemBlocks, when non-empty, overrides System and is sent as the
	// array form (with cache_control markers), matching claude-cli.
	SystemBlocks  []SystemBlock     `json:"-"`
	Tools         []ToolSchema      `json:"tools,omitempty"`
	Stream        bool              `json:"stream"`
	ThinkingBudget int             `json:"thinking_budget_tokens,omitempty"`
	// Thinking, when non-nil, is sent as the thinking field (adaptive/enabled).
	Thinking *ThinkingConfig `json:"thinking,omitempty"`
	// Metadata, when non-nil, is sent as the metadata field (user_id).
	Metadata *RequestMetadata `json:"metadata,omitempty"`
	// QuerySource marks the request source for retry policy (foreground/background).
	// Excluded from JSON serialisation.
	QuerySource string `json:"-"`
}

// MessageResponse corresponds to the non-streaming response body.
type MessageResponse struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Role       string         `json:"role"`
	Content    []ContentBlock `json:"content"`
	Model      string         `json:"model"`
	StopReason string         `json:"stop_reason"`
	Usage      Usage          `json:"usage"`
}

// Compile-time interface assertion: directClient must satisfy Client.
var _ Client = (*directClient)(nil)

// directClient implements Client using the Anthropic Direct API.
type directClient struct {
	apiKey     string
	authToken  string // ANTHROPIC_AUTH_TOKEN → Authorization: Bearer
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
}

func (c *directClient) Stream(ctx context.Context, req *MessageRequest) (StreamReader, error) {
	req.Stream = true
	httpReq, err := c.buildRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &APIError{Kind: ErrKindConnectionError, Message: err.Error()}
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, parseHTTPError(resp)
	}
	return newSSEReader(resp), nil
}

func (c *directClient) Complete(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	req.Stream = false
	httpReq, err := c.buildRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &APIError{Kind: ErrKindConnectionError, Message: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, parseHTTPError(resp)
	}
	var msg MessageResponse
	if decErr := json.NewDecoder(resp.Body).Decode(&msg); decErr != nil {
		return nil, fmt.Errorf("api: decode response: %w", decErr)
	}
	return &msg, nil
}

func (c *directClient) buildRequest(ctx context.Context, req *MessageRequest) (*http.Request, error) {
	// SystemBlocks（数组形式，含 cache_control）优先于 System 字符串。
	if len(req.SystemBlocks) > 0 {
		type shadowedRequest struct {
			MessageRequest
			System json.RawMessage `json:"system"`
		}
		blocks, err := json.Marshal(req.SystemBlocks)
		if err != nil {
			return nil, fmt.Errorf("api: marshal system blocks: %w", err)
		}
		sr := shadowedRequest{MessageRequest: *req, System: blocks}
		return c.marshalAndBuild(ctx, &sr)
	}
	return c.marshalAndBuild(ctx, req)
}

// marshalAndBuild serialises req and builds the HTTP request with auth headers.
func (c *directClient) marshalAndBuild(ctx context.Context, req any) (*http.Request, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("api: marshal request: %w", err)
	}
	url := c.baseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("api: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// 认证头对齐 claude-cli：ANTHROPIC_AUTH_TOKEN → Authorization: Bearer，
	// 否则 ANTHROPIC_API_KEY → x-api-key。两者都没有时不发送认证头。
	if c.authToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	} else if c.apiKey != "" {
		httpReq.Header.Set("x-api-key", c.apiKey)
	}
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("anthropic-beta", "claude-code-20250219,interleaved-thinking-2025-05-14,context-management-2025-06-27,prompt-caching-scope-2026-01-05,effort-2025-11-24")
	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}
	return httpReq, nil
}

// parseHTTPError reads an error response body and returns an *APIError.
func parseHTTPError(resp *http.Response) *APIError {
	body, _ := io.ReadAll(resp.Body)
	var errBody struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	msg := string(body)
	if json.Unmarshal(body, &errBody) == nil && errBody.Error.Message != "" {
		msg = errBody.Error.Message
	}
	kind := classifyError(resp.StatusCode, msg)
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    msg,
		Headers:    resp.Header,
		Kind:       kind,
	}
}
