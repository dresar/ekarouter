package openai

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

type AdapterMode string

const (
	ModeStandard AdapterMode = "openai"
	ModeCodex    AdapterMode = "codex"
)

type Adapter struct {
	client *http.Client
	mode   AdapterMode
}

func NewAdapter(client *http.Client) *Adapter {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	return &Adapter{
		client: client,
		mode:   ModeStandard,
	}
}

func NewCodexAdapter(client *http.Client) *Adapter {
	a := NewAdapter(client)
	a.mode = ModeCodex
	return a
}

func (a *Adapter) Kind() string {
	return string(a.mode)
}

func (a *Adapter) clientFor(creds *providers.Credentials) *http.Client {
	if creds != nil && creds.HTTPClient != nil {
		return creds.HTTPClient
	}
	return a.client
}

func (a *Adapter) Models(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error) {
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
		return nil, &providers.ProviderError{StatusCode: 0, Class: providers.ErrorClassNetwork, Message: err.Error(), Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, providers.ClassifyHTTPError(resp.StatusCode, string(b))
	}

	var data struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var models []providers.ModelInfo
	for _, m := range data.Data {
		models = append(models, providers.ModelInfo{
			ID:           m.ID,
			Name:         m.ID,
			ContextLimit: 128000,
			Streaming:    true,
			Capabilities: ModelCapabilities(m.ID),
		})
	}
	if len(models) == 0 {
		return DefaultModels(), nil
	}
	return models, nil
}

func (a *Adapter) buildEndpointURL(creds *providers.Credentials) string {
	baseURL := "https://api.openai.com/v1"
	if creds != nil && creds.BaseURL != "" {
		baseURL = strings.TrimRight(creds.BaseURL, "/")
	}

	if a.mode == ModeCodex {
		if strings.HasSuffix(baseURL, "/responses") {
			return baseURL
		}
		return baseURL + "/responses"
	}
	if strings.HasSuffix(baseURL, "/chat/completions") {
		return baseURL
	}
	return baseURL + "/chat/completions"
}

func (a *Adapter) BuildRequestBody(req *providers.Request, stream bool) map[string]any {
	if a.mode == ModeCodex {
		input := BuildCodexInput(req.Messages)
		bodyData := map[string]any{
			"model":  req.Model,
			"input":  input,
			"stream": stream,
			"store":  false,
		}
		if req.Temperature != nil {
			bodyData["temperature"] = *req.Temperature
		}
		if len(req.Tools) > 0 {
			bodyData["tools"] = NormalizeCodexTools(req.Tools)
		}
		if req.ToolChoice != nil {
			bodyData["tool_choice"] = req.ToolChoice
		}
		if req.MaxTokens != nil && *req.MaxTokens > 0 {
			bodyData["max_output_tokens"] = *req.MaxTokens
		}
		return bodyData
	}

	isReasoning := IsReasoningModel(req.Model)
	messages := AdaptMessagesForReasoning(req.Messages, isReasoning)

	bodyData := map[string]any{
		"model":    req.Model,
		"messages": messages,
		"stream":   stream,
	}

	if req.Temperature != nil && !isReasoning {
		bodyData["temperature"] = *req.Temperature
	}
	if req.TopP != nil && !isReasoning {
		bodyData["top_p"] = *req.TopP
	}
	if req.MaxTokens != nil {
		bodyData["max_tokens"] = *req.MaxTokens
	}
	if len(req.Stop) > 0 {
		bodyData["stop"] = req.Stop
	}
	if len(req.Tools) > 0 {
		bodyData["tools"] = req.Tools
	}
	if req.ToolChoice != nil {
		bodyData["tool_choice"] = req.ToolChoice
	}
	if req.ResponseFormat != nil {
		bodyData["response_format"] = req.ResponseFormat
	}

	ApplyReasoningParameters(bodyData, req, isReasoning)

	return bodyData
}

func (a *Adapter) prepareHTTPRequest(ctx context.Context, req *providers.Request, creds *providers.Credentials, stream bool) (*http.Request, error) {
	url := a.buildEndpointURL(creds)
	bodyData := a.BuildRequestBody(req, stream)

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	}

	if creds != nil {
		if creds.APIKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+creds.APIKey)
		} else if creds.AccessToken != "" {
			httpReq.Header.Set("Authorization", "Bearer "+creds.AccessToken)
		}
		for k, v := range creds.Headers {
			httpReq.Header.Set(k, v)
		}
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

	if a.mode == ModeCodex {
		codexResp, err := ParseResponsesResponse(rawBytes)
		if err == nil && (len(codexResp.Output) > 0 || codexResp.Status != "") {
			res := &providers.Response{
				ID:           codexResp.ID,
				Model:        codexResp.Model,
				Role:         "assistant",
				Content:      ExtractResponsesOutputText(codexResp.Output),
				FinishReason: codexResp.Status,
			}
			if codexResp.Usage != nil {
				u := ExtractResponsesUsage(codexResp.Usage)
				if u != nil {
					res.Usage = *u
				}
			}
			return res, nil
		}
	}

	openAIResp, err := ParseChatResponse(rawBytes)
	if err != nil {
		return nil, err
	}

	res := &providers.Response{
		ID:    openAIResp.ID,
		Model: openAIResp.Model,
	}

	if openAIResp.Usage != nil {
		u := ConvertUsage(openAIResp.Usage)
		if u != nil {
			res.Usage = *u
		}
	}

	if len(openAIResp.Choices) > 0 {
		c := openAIResp.Choices[0]
		res.Role = c.Message.Role
		res.Content = c.Message.Content
		res.Reasoning = c.Message.ReasoningContent
		res.FinishReason = c.FinishReason
	}

	return res, nil
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
			if strings.TrimSpace(data) == "[DONE]" {
				select {
				case events <- providers.StreamEvent{Type: providers.StreamEventDone}:
				case <-ctx.Done():
				}
				return
			}

			if a.mode == ModeCodex {
				codexChunk, err := ParseResponsesChunk([]byte(data))
				if err == nil && codexChunk.Delta != "" {
					select {
					case events <- providers.StreamEvent{
						Type:  providers.StreamEventDelta,
						Delta: codexChunk.Delta,
					}:
					case <-ctx.Done():
						return
					}
					continue
				}
			}

			chunk, err := ParseChatChunk([]byte(data))
			if err != nil {
				continue
			}

			if chunk.Usage != nil {
				u := ConvertUsage(chunk.Usage)
				if u != nil {
					select {
					case events <- providers.StreamEvent{
						Type:  providers.StreamEventUsage,
						Usage: u,
					}:
					case <-ctx.Done():
						return
					}
				}
			}

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta.ReasoningContent != "" {
					select {
					case events <- providers.StreamEvent{
						Type:      providers.StreamEventReasoning,
						Reasoning: delta.ReasoningContent,
					}:
					case <-ctx.Done():
						return
					}
				}
				if delta.Content != "" {
					select {
					case events <- providers.StreamEvent{
						Type:  providers.StreamEventDelta,
						Delta: delta.Content,
					}:
					case <-ctx.Done():
						return
					}
				}
			}
		}

		if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
			select {
			case events <- providers.StreamEvent{Type: providers.StreamEventError, Error: err}:
			case <-ctx.Done():
			}
			return
		}

		select {
		case events <- providers.StreamEvent{Type: providers.StreamEventDone}:
		case <-ctx.Done():
		}
	}()

	return events, nil
}
