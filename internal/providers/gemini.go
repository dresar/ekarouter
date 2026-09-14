package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GeminiAdapter struct {
	client *http.Client
}

func NewGeminiAdapter(client *http.Client) *GeminiAdapter {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	return &GeminiAdapter{client: client}
}

func (a *GeminiAdapter) Kind() string {
	return "gemini"
}

func (a *GeminiAdapter) Models(ctx context.Context, creds *Credentials) ([]ModelInfo, error) {
	return []ModelInfo{
		{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", ContextLimit: 1048576, Streaming: true},
		{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", ContextLimit: 2097152, Streaming: true},
		{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", ContextLimit: 1048576, Streaming: true},
	}, nil
}

func (a *GeminiAdapter) Execute(ctx context.Context, req *Request, creds *Credentials) (*Response, error) {
	baseURL := "https://generativelanguage.googleapis.com/v1beta/models"
	if creds != nil && creds.BaseURL != "" {
		baseURL = strings.TrimRight(creds.BaseURL, "/")
	}

	apiKey := ""
	if creds != nil {
		apiKey = creds.APIKey
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", baseURL, req.Model, apiKey)
	bodyData := a.buildRequestBody(req)
	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, &ProviderError{StatusCode: 0, Class: ErrorClassNetwork, Message: err.Error(), Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, ClassifyHTTPError(resp.StatusCode, string(b))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Role  string `json:"role"`
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, err
	}

	res := &Response{
		Model: req.Model,
		Usage: Usage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		},
	}

	if len(geminiResp.Candidates) > 0 {
		cand := geminiResp.Candidates[0]
		res.Role = cand.Content.Role
		res.FinishReason = cand.FinishReason
		var text strings.Builder
		for _, p := range cand.Content.Parts {
			text.WriteString(p.Text)
		}
		res.Content = text.String()
	}

	return res, nil
}

func (a *GeminiAdapter) ExecuteStream(ctx context.Context, req *Request, creds *Credentials) (<-chan StreamEvent, error) {
	baseURL := "https://generativelanguage.googleapis.com/v1beta/models"
	if creds != nil && creds.BaseURL != "" {
		baseURL = strings.TrimRight(creds.BaseURL, "/")
	}

	apiKey := ""
	if creds != nil {
		apiKey = creds.APIKey
	}

	url := fmt.Sprintf("%s/%s:streamGenerateContent?alt=sse&key=%s", baseURL, req.Model, apiKey)
	bodyData := a.buildRequestBody(req)
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
			var chunk struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
				} `json:"candidates"`
				UsageMetadata *struct {
					PromptTokenCount     int `json:"promptTokenCount"`
					CandidatesTokenCount int `json:"candidatesTokenCount"`
					TotalTokenCount      int `json:"totalTokenCount"`
				} `json:"usageMetadata,omitempty"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if chunk.UsageMetadata != nil {
				events <- StreamEvent{
					Type: StreamEventUsage,
					Usage: &Usage{
						PromptTokens:     chunk.UsageMetadata.PromptTokenCount,
						CompletionTokens: chunk.UsageMetadata.CandidatesTokenCount,
						TotalTokens:      chunk.UsageMetadata.TotalTokenCount,
					},
				}
			}

			if len(chunk.Candidates) > 0 {
				for _, part := range chunk.Candidates[0].Content.Parts {
					if part.Text != "" {
						events <- StreamEvent{
							Type:  StreamEventDelta,
							Delta: part.Text,
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
			events <- StreamEvent{Type: StreamEventError, Error: err}
			return
		}

		events <- StreamEvent{Type: StreamEventDone}
	}()

	return events, nil
}

func (a *GeminiAdapter) buildRequestBody(req *Request) map[string]any {
	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	}

	var contents []content
	for _, m := range req.Messages {
		role := "user"
		if strings.ToLower(m.Role) == "assistant" {
			role = "model"
		}
		contents = append(contents, content{
			Role:  role,
			Parts: []part{{Text: m.Content}},
		})
	}

	genConfig := make(map[string]any)
	if req.Temperature != nil {
		genConfig["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		genConfig["topP"] = *req.TopP
	}
	if req.MaxTokens != nil {
		genConfig["maxOutputTokens"] = *req.MaxTokens
	}

	body := map[string]any{
		"contents": contents,
	}
	if len(genConfig) > 0 {
		body["generationConfig"] = genConfig
	}

	return body
}
