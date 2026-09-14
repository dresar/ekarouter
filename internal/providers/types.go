package providers

import (
	"context"
	"net/http"
)

type ModelInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ContextLimit int    `json:"context_limit"`
	Streaming    bool   `json:"streaming"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Request struct {
	ID               string    `json:"id"`
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Temperature      *float64  `json:"temperature,omitempty"`
	TopP             *float64  `json:"top_p,omitempty"`
	MaxTokens        *int      `json:"max_tokens,omitempty"`
	Stream           bool      `json:"stream"`
	Stop             []string  `json:"stop,omitempty"`
	OptOutTokenSaver bool      `json:"opt_out_token_saver,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Response struct {
	ID           string `json:"id"`
	Model        string `json:"model"`
	Role         string `json:"role"`
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
	Usage        Usage  `json:"usage"`
}

type StreamEventType string

const (
	StreamEventDelta StreamEventType = "delta"
	StreamEventUsage StreamEventType = "usage"
	StreamEventDone  StreamEventType = "done"
	StreamEventError StreamEventType = "error"
)

type StreamEvent struct {
	Type  StreamEventType
	Delta string
	Usage *Usage
	Error error
}

type Credentials struct {
	APIKey      string
	AccessToken string
	SecretKey   string
	BaseURL     string
	HTTPClient  *http.Client
}

type Adapter interface {
	Kind() string
	Models(ctx context.Context, creds *Credentials) ([]ModelInfo, error)
	Execute(ctx context.Context, req *Request, creds *Credentials) (*Response, error)
	ExecuteStream(ctx context.Context, req *Request, creds *Credentials) (<-chan StreamEvent, error)
}
