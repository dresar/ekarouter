package anthropic

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

	"github.com/dresar/ekarouter/internal/providers"
)

type Adapter struct {
	client *http.Client
}

func NewAdapter(client *http.Client) *Adapter {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	return &Adapter{client: client}
}

func (a *Adapter) Kind() string {
	return "anthropic"
}

func (a *Adapter) clientFor(creds *providers.Credentials) *http.Client {
	if creds != nil && creds.HTTPClient != nil {
		return creds.HTTPClient
	}
	return a.client
}

func (a *Adapter) Models(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error) {
	return DefaultModels(), nil
}

func (a *Adapter) buildEndpointURL(creds *providers.Credentials) string {
	baseURL := "https://api.anthropic.com/v1"
	if creds != nil && creds.BaseURL != "" {
		baseURL = strings.TrimRight(creds.BaseURL, "/")
	}
	if strings.HasSuffix(baseURL, "/messages") {
		return baseURL
	}
	return baseURL + "/messages"
}

func (a *Adapter) BuildRequestBody(req *providers.Request, creds *providers.Credentials, stream bool) (map[string]any, bool) {
	systemPrompt := ConcatenateSystemPrompts(req.Messages)
	var nonSystemMsgs []map[string]any

	for _, m := range req.Messages {
		role := strings.ToLower(m.Role)
		if role == "system" {
			continue
		}
		if role != "assistant" {
			role = "user"
		}
		nonSystemMsgs = append(nonSystemMsgs, map[string]any{
			"role":    role,
			"content": m.Content,
		})
	}

	if len(nonSystemMsgs) == 0 {
		nonSystemMsgs = append(nonSystemMsgs, map[string]any{
			"role":    "user",
			"content": "Hello",
		})
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
	if len(req.Stop) > 0 {
		bodyData["stop_sequences"] = req.Stop
	}
	if len(req.Tools) > 0 {
		if creds != nil && creds.AccessToken != "" {
			bodyData["tools"] = CloakTools(req.Tools)
		} else {
			bodyData["tools"] = req.Tools
		}
	}

	hasThinking := ConfigureThinking(bodyData, req)

	return bodyData, hasThinking
}

func (a *Adapter) prepareHTTPRequest(ctx context.Context, req *providers.Request, creds *providers.Credentials, stream bool) (*http.Request, error) {
	url := a.buildEndpointURL(creds)
	bodyData, hasThinking := a.BuildRequestBody(req, creds, stream)

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	ApplyAnthropicHeaders(httpReq.Header, creds, hasThinking)
	if stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	}

	return httpReq, nil
}

func (a *Adapter) Execute(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
	httpReq, err := a.prepareHTTPRequest(ctx, req, creds, false)
	if err != nil {
		return nil, err
	}

	resp, err := a.clientFor(creds).Do(httpReq)
	if err != nil {
		return nil, &providers.ProviderError{StatusCode: 0, Class: providers.ErrorClassNetwork, Message: err.Error(), Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, providers.ClassifyHTTPError(resp.StatusCode, string(b))
	}

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	anthropicResp, err := ParseMessageResponse(rawBytes)
	if err != nil {
		return nil, err
	}

	text, thinking := ExtractBlocks(anthropicResp.Content)
	usage := ConvertUsage(anthropicResp.Usage.InputTokens, anthropicResp.Usage.OutputTokens)

	return &providers.Response{
		ID:           anthropicResp.ID,
		Model:        anthropicResp.Model,
		Role:         anthropicResp.Role,
		Content:      text,
		Reasoning:    thinking,
		FinishReason: anthropicResp.StopReason,
		Usage:        usage,
	}, nil
}

func (a *Adapter) ExecuteStream(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
	httpReq, err := a.prepareHTTPRequest(ctx, req, creds, true)
	if err != nil {
		return nil, err
	}

	resp, err := a.clientFor(creds).Do(httpReq)
	if err != nil {
		return nil, &providers.ProviderError{StatusCode: 0, Class: providers.ErrorClassNetwork, Message: err.Error(), Err: err}
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, providers.ClassifyHTTPError(resp.StatusCode, string(b))
	}

	events := make(chan providers.StreamEvent, 16)

	go func() {
		defer resp.Body.Close()
		defer close(events)

		scanner := bufio.NewScanner(resp.Body)
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 4*1024*1024)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				select {
				case events <- providers.StreamEvent{Type: providers.StreamEventError, Error: ctx.Err()}:
				default:
				}
				return
			default:
			}

			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			eventData, err := ParseStreamEventPayload([]byte(data))
			if err != nil {
				continue
			}

			switch eventData.Type {
			case "content_block_delta":
				if eventData.Delta.Thinking != "" {
					select {
					case events <- providers.StreamEvent{
						Type:      providers.StreamEventReasoning,
						Reasoning: eventData.Delta.Thinking,
					}:
					case <-ctx.Done():
						return
					}
				} else if eventData.Delta.Text != "" {
					select {
					case events <- providers.StreamEvent{
						Type:  providers.StreamEventDelta,
						Delta: eventData.Delta.Text,
					}:
					case <-ctx.Done():
						return
					}
				}
			case "message_delta":
				if eventData.Usage != nil {
					select {
					case events <- providers.StreamEvent{
						Type: providers.StreamEventUsage,
						Usage: &providers.Usage{
							CompletionTokens: eventData.Usage.OutputTokens,
						},
					}:
					case <-ctx.Done():
						return
					}
				}
			case "message_stop":
				select {
				case events <- providers.StreamEvent{Type: providers.StreamEventDone}:
				case <-ctx.Done():
				}
				return
			}
		}

		if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
			select {
			case events <- providers.StreamEvent{Type: providers.StreamEventError, Error: err}:
			case <-ctx.Done():
			}
		}
	}()

	return events, nil
}
