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

	isImage := IsImageModel(req.Model)
	switch a.mode {
	case ModeAntigravity:
		return BuildAntigravityURL(baseURL, stream, isImage)
	case ModeCLI:
		return BuildCLIURL(baseURL, req.Model, stream)
	default:
		cleaned := strings.TrimRight(baseURL, "/")
		if cleaned == "" {
			cleaned = "https://generativelanguage.googleapis.com/v1beta/models"
		}
		if stream && !isImage {
			return fmt.Sprintf("%s/%s:streamGenerateContent?alt=sse&key=%s", cleaned, req.Model, apiKey)
		}
		return fmt.Sprintf("%s/%s:generateContent?key=%s", cleaned, req.Model, apiKey)
	}
}

func (a *Adapter) BuildRequestBody(req *providers.Request, creds *providers.Credentials) map[string]any {
	var contents []map[string]any
	var systemParts []map[string]any

	for _, m := range req.Messages {
		r := strings.ToLower(m.Role)
		if r == "system" {
			systemParts = append(systemParts, map[string]any{"text": m.Content})
			continue
		}
		role := "user"
		if r == "assistant" || r == "model" {
			role = "model"
		}

		var parts []map[string]any
		if m.Content != "" {
			parts = append(parts, map[string]any{"text": m.Content})
		}

		if len(m.ToolCalls) > 0 {
			firstCall := true
			for _, tc := range m.ToolCalls {
				var args map[string]any
				if tc.Function.Arguments != "" {
					_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
				}
				if args == nil {
					args = make(map[string]any)
				}
				fCall := map[string]any{
					"name": SanitizeToolName(tc.Function.Name),
					"args": args,
				}
				p := map[string]any{"functionCall": fCall}
				if a.mode == ModeAntigravity && firstCall {
					sig := ""
					if tc.ID != "" && a.signatures != nil {
						sig = a.signatures.Get(tc.ID)
					}
					if sig == "" && req.ID != "" && a.signatures != nil {
						sig = a.signatures.Get(req.ID)
					}
					if sig == "" {
						sig = DefaultThinkingAGSignature
					}
					p["thoughtSignature"] = sig
					firstCall = false
				}
				parts = append(parts, p)
			}
		}

		if r == "tool" {
			role = "user"
			var respObj map[string]any
			if err := json.Unmarshal([]byte(m.Content), &respObj); err != nil {
				respObj = map[string]any{"result": m.Content}
			}
			parts = append(parts, map[string]any{
				"functionResponse": map[string]any{
					"name":     SanitizeToolName(m.Name),
					"response": respObj,
				},
			})
		}

		if len(parts) == 0 {
			parts = append(parts, map[string]any{"text": ""})
		}

		contents = append(contents, map[string]any{
			"role":  role,
			"parts": parts,
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
	if IsImageModel(req.Model) {
		promptConcat := ""
		for _, m := range req.Messages {
			promptConcat += " " + m.Content
		}
		genConfig["aspectRatio"] = ExtractAspectRatio(promptConcat)
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

	if len(req.Tools) > 0 {
		var decls []map[string]any
		for _, t := range req.Tools {
			if tm, ok := t.(map[string]any); ok {
				if fn, ok := tm["function"].(map[string]any); ok {
					name, _ := fn["name"].(string)
					desc, _ := fn["description"].(string)
					params := fn["parameters"]
					decls = append(decls, map[string]any{
						"name":        SanitizeToolName(name),
						"description": desc,
						"parameters":  params,
					})
				}
			}
		}
		if len(decls) > 0 {
			coreBody["tools"] = []map[string]any{
				{"functionDeclarations": decls},
			}
			if a.mode == ModeAntigravity {
				coreBody["toolConfig"] = map[string]any{
					"functionCallingConfig": map[string]any{
						"mode": "VALIDATED",
					},
				}
			}
		}
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
		for _, part := range cand.Content.Parts {
			if part.ThoughtSignature != "" && a.signatures != nil {
				if part.FunctionCall != nil && part.FunctionCall.Name != "" {
					a.signatures.Store(part.FunctionCall.Name, part.ThoughtSignature)
				}
				if req.ID != "" {
					a.signatures.Store(req.ID, part.ThoughtSignature)
				}
			}
		}
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
					if part.ThoughtSignature != "" && a.signatures != nil {
						if part.FunctionCall != nil && part.FunctionCall.Name != "" {
							a.signatures.Store(part.FunctionCall.Name, part.ThoughtSignature)
						}
						if req.ID != "" {
							a.signatures.Store(req.ID, part.ThoughtSignature)
						}
					}
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
