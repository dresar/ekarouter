package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

type AnthropicAdapter struct {
	client *http.Client
}

func NewAnthropicAdapter(client *http.Client) *AnthropicAdapter {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	return &AnthropicAdapter{client: client}
}

func (a *AnthropicAdapter) Kind() string {
	return "anthropic"
}

func (a *AnthropicAdapter) Models(ctx context.Context, creds *Credentials) ([]ModelInfo, error) {
	return []ModelInfo{
		{ID: "claude-3-7-sonnet-20250219", Name: "Claude 3.7 Sonnet", ContextLimit: 200000, Streaming: true},
		{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", ContextLimit: 200000, Streaming: true},
		{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", ContextLimit: 200000, Streaming: true},
	}, nil
}

func (a *AnthropicAdapter) Execute(ctx context.Context, req *Request, creds *Credentials) (*Response, error) {
	url := "https://api.anthropic.com/v1/messages"
	if creds != nil && creds.BaseURL != "" {
		url = strings.TrimRight(creds.BaseURL, "/") + "/messages"
	}

	bodyData := a.buildRequestBody(req, false)
	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	if creds != nil && creds.APIKey != "" {
		httpReq.Header.Set("x-api-key", creds.APIKey)
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, &ProviderError{StatusCode: 0, Class: ErrorClassNetwork, Message: err.Error(), Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, ClassifyHTTPError(resp.StatusCode, string(b))
	}

	var anthropicResp struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, err
	}

	var fullContent strings.Builder
	for _, c := range anthropicResp.Content {
		if c.Type == "text" {
			fullContent.WriteString(c.Text)
		}
	}

	return &Response{
		ID:           anthropicResp.ID,
		Model:        anthropicResp.Model,
		Role:         anthropicResp.Role,
		Content:      fullContent.String(),
		FinishReason: anthropicResp.StopReason,
		Usage: Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}, nil
}

func (a *AnthropicAdapter) ExecuteStream(ctx context.Context, req *Request, creds *Credentials) (<-chan StreamEvent, error) {
	url := "https://api.anthropic.com/v1/messages"
	if creds != nil && creds.BaseURL != "" {
		url = strings.TrimRight(creds.BaseURL, "/") + "/messages"
	}

	bodyData := a.buildRequestBody(req, true)
	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")
	if creds != nil && creds.APIKey != "" {
		httpReq.Header.Set("x-api-key", creds.APIKey)
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, &ProviderError{StatusCode: 0, Class: ErrorClassNetwork, Message: err.Error(), Err: err}
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, ClassifyHTTPError(resp.StatusCode, string(b))
	}

	events := make(chan StreamEvent, 16)

	go func() {
		defer resp.Body.Close()
		defer close(events)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				events <- StreamEvent{Type: StreamEventError, Error: ctx.Err()}
				return
			default:
			}

			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			var eventData struct {
				Type  string `json:"type"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
				Usage *struct {
					OutputTokens int `json:"output_tokens"`
				} `json:"usage,omitempty"`
			}

			if err := json.Unmarshal([]byte(data), &eventData); err != nil {
				continue
			}

			switch eventData.Type {
			case "content_block_delta":
				if eventData.Delta.Text != "" {
					events <- StreamEvent{
						Type:  StreamEventDelta,
						Delta: eventData.Delta.Text,
					}
				}
			case "message_delta":
				if eventData.Usage != nil {
					events <- StreamEvent{
						Type: StreamEventUsage,
						Usage: &Usage{
							CompletionTokens: eventData.Usage.OutputTokens,
						},
					}
				}
			case "message_stop":
				events <- StreamEvent{Type: StreamEventDone}
				return
			}
		}

		if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
			events <- StreamEvent{Type: StreamEventError, Error: err}
		}
	}()

	return events, nil
}

func (a *AnthropicAdapter) buildRequestBody(req *Request, stream bool) map[string]any {
	var systemPrompt string
	var nonSystemMsgs []map[string]string

	for _, m := range req.Messages {
		if strings.ToLower(m.Role) == "system" {
			systemPrompt = m.Content
		} else {
			role := "user"
			if strings.ToLower(m.Role) == "assistant" {
				role = "assistant"
			}
			nonSystemMsgs = append(nonSystemMsgs, map[string]string{
				"role":    role,
				"content": m.Content,
			})
		}
	}

	maxTokens := 4096
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		maxTokens = *req.MaxTokens
	}

	bodyData := map[string]any{
		"model":      req.Model,
		"messages":   nonSystemMsgs,
		"max_tokens": maxTokens,
		"stream":     stream,
	}
	if systemPrompt != "" {
		bodyData["system"] = systemPrompt
	}
	if req.Temperature != nil {
		bodyData["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		bodyData["top_p"] = *req.TopP
	}

	return bodyData
}
