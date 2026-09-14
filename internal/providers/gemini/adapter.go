package gemini

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

	"github.com/dresar/ekarouter/internal/providers"
)

type AdapterMode string

const (
	ModeStandard    AdapterMode = "gemini"
	ModeCLI         AdapterMode = "gemini-cli"
	ModeAntigravity AdapterMode = "antigravity"
)

type Adapter struct {
	client     *http.Client
	mode       AdapterMode
	signatures *ThoughtSignatureStore
}

func NewAdapter(client *http.Client) *Adapter {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	return &Adapter{
		client:     client,
		mode:       ModeStandard,
		signatures: NewThoughtSignatureStore(),
	}
}

func NewCLIAdapter(client *http.Client) *Adapter {
	a := NewAdapter(client)
	a.mode = ModeCLI
	return a
}

func NewAntigravityAdapter(client *http.Client) *Adapter {
	a := NewAdapter(client)
	a.mode = ModeAntigravity
	return a
}

func (a *Adapter) Kind() string {
	return string(a.mode)
}

func (a *Adapter) ThoughtSignatures() *ThoughtSignatureStore {
	return a.signatures
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

func (a *Adapter) buildEndpointURL(req *providers.Request, creds *providers.Credentials, stream bool) string {
	baseURL := ""
	if creds != nil && creds.BaseURL != "" {
		baseURL = creds.BaseURL
	}

	apiKey := ""
	if creds != nil {
		apiKey = creds.APIKey
	}

	switch a.mode {
	case ModeAntigravity:
		return BuildAntigravityURL(baseURL)
	case ModeCLI:
		return BuildCLIURL(baseURL, req.Model)
	default:
		cleaned := strings.TrimRight(baseURL, "/")
		if cleaned == "" {
			cleaned = "https://generativelanguage.googleapis.com/v1beta/models"
		}
		if stream {
			return fmt.Sprintf("%s/%s:streamGenerateContent?alt=sse&key=%s", cleaned, req.Model, apiKey)
		}
		return fmt.Sprintf("%s/%s:generateContent?key=%s", cleaned, req.Model, apiKey)
	}
}

func (a *Adapter) BuildRequestBody(req *providers.Request, creds *providers.Credentials) map[string]any {
	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	}

	var contents []content
	var systemParts []part

	for _, m := range req.Messages {
		r := strings.ToLower(m.Role)
		if r == "system" {
			systemParts = append(systemParts, part{Text: m.Content})
			continue
		}
		role := "user"
		if r == "assistant" || r == "model" {
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

	coreBody := map[string]any{
		"contents": contents,
	}
	if len(systemParts) > 0 {
		coreBody["systemInstruction"] = map[string]any{
			"parts": systemParts,
		}
	}
	if len(genConfig) > 0 {
		coreBody["generationConfig"] = genConfig
	}

	switch a.mode {
	case ModeAntigravity:
		projectID := ResolveProjectID(req, creds, "antigravity-cloudcode-project")
		return WrapAntigravityRequest(projectID, req.Model, coreBody)
	case ModeCLI:
		projectID := ResolveProjectID(req, creds, "gemini-cli-project")
		return WrapCLIRequest(projectID, req.Model, coreBody)
	default:
		return coreBody
	}
}

func (a *Adapter) prepareHTTPRequest(ctx context.Context, req *providers.Request, creds *providers.Credentials, stream bool) (*http.Request, error) {
	url := a.buildEndpointURL(req, creds, stream)
	bodyData := a.BuildRequestBody(req, creds)

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	switch a.mode {
	case ModeAntigravity:
		h := BuildAntigravityHeaders(creds)
		for k, v := range h {
			httpReq.Header[k] = v
		}
	case ModeCLI:
		h := BuildCLIHeaders(creds)
		for k, v := range h {
			httpReq.Header[k] = v
		}
	default:
		httpReq.Header.Set("Content-Type", "application/json")
		if creds != nil && creds.AccessToken != "" {
			httpReq.Header.Set("Authorization", "Bearer "+creds.AccessToken)
		}
	}

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

	payload, err := ParseResponsePayload(rawBytes)
	if err != nil {
		return nil, err
	}

	res := &providers.Response{
		Model: req.Model,
	}

	if payload.UsageMetadata != nil {
		u := ExtractUsage(payload.UsageMetadata)
		if u != nil {
			res.Usage = *u
		}
	}

	if len(payload.Candidates) > 0 {
		cand := payload.Candidates[0]
		res.Role = cand.Content.Role
		res.FinishReason = NormalizeFinishReason(cand.FinishReason)
		text, reasoning := ExtractCandidateTextAndReasoning(cand)
		res.Content = text
		res.Reasoning = reasoning
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
			payload, err := ParseResponsePayload([]byte(data))
			if err != nil {
				continue
			}

			if payload.UsageMetadata != nil {
				u := ExtractUsage(payload.UsageMetadata)
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

			if len(payload.Candidates) > 0 {
				cand := payload.Candidates[0]
				for _, part := range cand.Content.Parts {
					if part.Thought && part.Text != "" {
						select {
						case events <- providers.StreamEvent{
							Type:      providers.StreamEventReasoning,
							Reasoning: part.Text,
						}:
						case <-ctx.Done():
							return
						}
					} else if part.Text != "" {
						select {
						case events <- providers.StreamEvent{
							Type:  providers.StreamEventDelta,
							Delta: part.Text,
						}:
						case <-ctx.Done():
							return
						}
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
