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

type OpenAIAdapter struct {
	client *http.Client
}

func NewOpenAIAdapter(client *http.Client) *OpenAIAdapter {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	return &OpenAIAdapter{client: client}
}

func (a *OpenAIAdapter) Kind() string {
	return "openai"
}

func (a *OpenAIAdapter) clientFor(creds *Credentials) *http.Client {
	if creds != nil && creds.HTTPClient != nil {
		return creds.HTTPClient
	}
	return a.client
}

func (a *OpenAIAdapter) Models(ctx context.Context, creds *Credentials) ([]ModelInfo, error) {
	baseURL := "https://api.openai.com/v1"
	if creds != nil && creds.BaseURL != "" {
		baseURL = strings.TrimRight(creds.BaseURL, "/")
	}

	url := baseURL + "/models"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if creds != nil && creds.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+creds.APIKey)
	}

	resp, err := a.clientFor(creds).Do(httpReq)
	if err != nil {
		return nil, &ProviderError{StatusCode: 0, Class: ErrorClassNetwork, Message: err.Error(), Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, ClassifyHTTPError(resp.StatusCode, string(b))
	}

	var data struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var models []ModelInfo
	for _, m := range data.Data {
		models = append(models, ModelInfo{
			ID:           m.ID,
			Name:         m.ID,
			ContextLimit: 8192,
			Streaming:    true,
		})
	}
	return models, nil
}

func (a *OpenAIAdapter) Execute(ctx context.Context, req *Request, creds *Credentials) (*Response, error) {
	baseURL := "https://api.openai.com/v1"
	if creds != nil && creds.BaseURL != "" {
		baseURL = strings.TrimRight(creds.BaseURL, "/")
	}

	url := baseURL + "/chat/completions"

	bodyData := map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   false,
	}
	if req.Temperature != nil {
		bodyData["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		bodyData["top_p"] = *req.TopP
	}
	if req.MaxTokens != nil {
		bodyData["max_tokens"] = *req.MaxTokens
	}
	if len(req.Stop) > 0 {
		bodyData["stop"] = req.Stop
	}

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if creds != nil && creds.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+creds.APIKey)
	}

	resp, err := a.clientFor(creds).Do(httpReq)
	if err != nil {
		return nil, &ProviderError{StatusCode: 0, Class: ErrorClassNetwork, Message: err.Error(), Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, ClassifyHTTPError(resp.StatusCode, string(b))
	}

	var openAIResp struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	res := &Response{
		ID:    openAIResp.ID,
		Model: openAIResp.Model,
		Usage: Usage{
			PromptTokens:     openAIResp.Usage.PromptTokens,
			CompletionTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:      openAIResp.Usage.TotalTokens,
		},
	}
	if len(openAIResp.Choices) > 0 {
		res.Role = openAIResp.Choices[0].Message.Role
		res.Content = openAIResp.Choices[0].Message.Content
		res.FinishReason = openAIResp.Choices[0].FinishReason
	}
	return res, nil
}

func (a *OpenAIAdapter) ExecuteStream(ctx context.Context, req *Request, creds *Credentials) (<-chan StreamEvent, error) {
	baseURL := "https://api.openai.com/v1"
	if creds != nil && creds.BaseURL != "" {
		baseURL = strings.TrimRight(creds.BaseURL, "/")
	}

	url := baseURL + "/chat/completions"

	bodyData := map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	}
	if req.Temperature != nil {
		bodyData["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		bodyData["top_p"] = *req.TopP
	}
	if req.MaxTokens != nil {
		bodyData["max_tokens"] = *req.MaxTokens
	}

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if creds != nil && creds.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+creds.APIKey)
	}

	resp, err := a.clientFor(creds).Do(httpReq)
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
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 4*1024*1024)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				select {
				case events <- StreamEvent{Type: StreamEventError, Error: ctx.Err()}:
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
			if strings.TrimSpace(data) == "[DONE]" {
				select {
				case events <- StreamEvent{Type: StreamEventDone}:
				case <-ctx.Done():
				}
				return
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
				Usage *struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
					TotalTokens      int `json:"total_tokens"`
				} `json:"usage,omitempty"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if chunk.Usage != nil {
				select {
				case events <- StreamEvent{
					Type: StreamEventUsage,
					Usage: &Usage{
						PromptTokens:     chunk.Usage.PromptTokens,
						CompletionTokens: chunk.Usage.CompletionTokens,
						TotalTokens:      chunk.Usage.TotalTokens,
					},
				}:
				case <-ctx.Done():
					return
				}
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				select {
				case events <- StreamEvent{
					Type:  StreamEventDelta,
					Delta: chunk.Choices[0].Delta.Content,
				}:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
			select {
			case events <- StreamEvent{Type: StreamEventError, Error: err}:
			case <-ctx.Done():
			}
		}
	}()

	return events, nil
}
