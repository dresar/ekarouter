package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/gateway"
	"github.com/dresar/ekarouter/internal/httpapi"
	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/routing"
)

func Test9Router_ModelResolutionAndAliases(t *testing.T) {
	cases := []struct {
		input            string
		expectedProvider string
		expectedModel    string
	}{
		{"ag/gemini-2.5-flash", "antigravity", "gemini-2.5-flash"},
		{"ag/gemini-3.8-flash-high(high)", "antigravity", "gemini-3.8-flash-high"},
		{"oa/gpt-4o", "openai", "gpt-4o"},
		{"an/claude-3-7-sonnet", "anthropic", "claude-3-7-sonnet"},
		{"gm/gemini-2.5-pro", "gemini", "gemini-2.5-pro"},
		{"or/deepseek/deepseek-r1", "openrouter", "deepseek/deepseek-r1"},
		{"gr/llama-3.3-70b-versatile", "groq", "llama-3.3-70b-versatile"},
		{"kr/claude-sonnet-4-5", "kiro", "claude-sonnet-4-5"},
		{"oc/gpt-4o", "opencode", "gpt-4o"},
		{"gcli/grok-3", "grok-cli", "grok-3"},
		{"cx/gpt-4o-review", "codex", "gpt-4o-review"},
		{"qd/deepseek-v3", "qoder", "deepseek-v3"},
		{"cf/llama-3.1-8b", "cloudflare-ai", "llama-3.1-8b"},
		{"ol/nomic-embed-text", "ollama", "nomic-embed-text"},
		{"gpt-4o-mini", "openai", "gpt-4o-mini"},
		{"claude-3-5-sonnet", "anthropic", "claude-3-5-sonnet"},
		{"gemini-2.5-flash", "gemini", "gemini-2.5-flash"},
		{"llama-3.3-70b-instruct", "groq", "llama-3.3-70b-instruct"},
		{"deepseek-chat", "deepseek", "deepseek-chat"},
	}

	for _, tc := range cases {
		info := routing.ParseModel(tc.input)
		if info.Provider != tc.expectedProvider {
			t.Errorf("model '%s': expected provider '%s', got '%s'", tc.input, tc.expectedProvider, info.Provider)
		}
		if info.Model != tc.expectedModel {
			t.Errorf("model '%s': expected model '%s', got '%s'", tc.input, tc.expectedModel, info.Model)
		}
	}
}

func Test9Router_ThinkingSuffixExtraction(t *testing.T) {
	cases := []struct {
		input          string
		expectedBase   string
		expectedSuffix string
	}{
		{"gemini-3.8-flash(high)", "gemini-3.8-flash", "(high)"},
		{"claude-sonnet-4-6(thinking)", "claude-sonnet-4-6", "(thinking)"},
		{"ag/gemini-3.7-flash-tiered(low)", "gemini-3.7-flash-tiered", "(low)"},
		{"gpt-oss-120b(thought)", "gpt-oss-120b", "(thought)"},
		{"plain-model", "plain-model", ""},
	}

	for _, tc := range cases {
		info := routing.ParseModel(tc.input)
		if info.Model != tc.expectedBase {
			t.Errorf("input '%s': expected base '%s', got '%s'", tc.input, tc.expectedBase, info.Model)
		}
		if info.ThinkingSuffix != tc.expectedSuffix {
			t.Errorf("input '%s': expected suffix '%s', got '%s'", tc.input, tc.expectedSuffix, info.ThinkingSuffix)
		}
	}
}

func Test9Router_VersionDotDashNormalization(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"claude-sonnet-4-5", "claude-sonnet-4.5"},
		{"gemini-2-5-flash", "gemini-2.5-flash"},
		{"claude-3-7-sonnet", "claude-3.7-sonnet"},
	}

	for _, tc := range cases {
		got := routing.NormalizeVersion(tc.input)
		if got != tc.expected {
			t.Errorf("input '%s': expected '%s', got '%s'", tc.input, tc.expected, got)
		}
	}
}

func Test9Router_DynamicRoutingExecution(t *testing.T) {
	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)

	router.SetProvider("p-ag", "antigravity")
	router.SetProvider("p-oa", "openai")
	router.SetProvider("p-an", "anthropic")
	router.SetProvider("p-gr", "groq")

	router.SetAccount(AccountWrapper("acc-ag-1", "p-ag", 10))
	router.SetAccount(AccountWrapper("acc-ag-2", "p-ag", 20))
	router.SetAccount(AccountWrapper("acc-oa-1", "p-oa", 10))
	router.SetAccount(AccountWrapper("acc-an-1", "p-an", 10))
	router.SetAccount(AccountWrapper("acc-gr-1", "p-gr", 10))

	targetsAg, err := router.SelectTargets("ag/gemini-2.5-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targetsAg) != 2 || targetsAg[0].AccountID != "acc-ag-1" {
		t.Fatalf("expected acc-ag-1 with priority 10 first, got: %+v", targetsAg)
	}

	cd.MarkFailure("acc-ag-1", 10*time.Second)

	targetsAgAfterCool, err := router.SelectTargets("ag/gemini-2.5-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targetsAgAfterCool) != 1 || targetsAgAfterCool[0].AccountID != "acc-ag-2" {
		t.Fatalf("expected acc-ag-2 after acc-ag-1 cooling down, got: %+v", targetsAgAfterCool)
	}

	targetsOa, err := router.SelectTargets("gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targetsOa) != 1 || targetsOa[0].AccountID != "acc-oa-1" {
		t.Fatalf("expected acc-oa-1 for gpt-4o, got: %+v", targetsOa)
	}

	targetsAn, err := router.SelectTargets("claude-3-7-sonnet(high)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targetsAn) != 1 || targetsAn[0].AccountID != "acc-an-1" {
		t.Fatalf("expected acc-an-1 for claude-3-7-sonnet(high), got: %+v", targetsAn)
	}
	if targetsAn[0].ModelName != "claude-3-7-sonnet(high)" {
		t.Fatalf("expected thinking suffix preserved in target model, got: %s", targetsAn[0].ModelName)
	}
}

func AccountWrapper(id, providerID string, priority int) *routing.Account {
	return &routing.Account{
		ID:         id,
		ProviderID: providerID,
		State:      "active",
		Enabled:    true,
		Priority:   priority,
	}
}

func Test9Router_EmbeddingsEndpoint(t *testing.T) {
	h := &httpapi.GatewayHandler{}

	t.Run("single string standard input", func(t *testing.T) {
		body := `{"model":"text-embedding-3-small","input":"9Router multi-provider test"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.Embeddings(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp httpapi.EmbeddingsResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed unmarshal: %v", err)
		}

		if resp.Object != "list" || len(resp.Data) != 1 || resp.Model != "text-embedding-3-small" {
			t.Errorf("unexpected response structure: %+v", resp)
		}
	})

	t.Run("batch array input", func(t *testing.T) {
		body := `{"model":"text-embedding-3-small","input":["chunk 1","chunk 2","chunk 3","chunk 4"]}`
		req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.Embeddings(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp httpapi.EmbeddingsResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)

		if len(resp.Data) != 4 {
			t.Fatalf("expected 4 embeddings, got %d", len(resp.Data))
		}
	})

	t.Run("base64 encoding format", func(t *testing.T) {
		body := `{"model":"text-embedding-3-small","input":"base64 test","encoding_format":"base64"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.Embeddings(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp httpapi.EmbeddingsResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)

		_, ok := resp.Data[0].Embedding.(string)
		if !ok {
			t.Fatalf("expected base64 string, got %T", resp.Data[0].Embedding)
		}
	})

	t.Run("validation error on empty input", func(t *testing.T) {
		body := `{"model":"text-embedding-3-small","input":""}`
		req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.Embeddings(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
		}
	})
}

func Test9Router_CanonicalModelsList(t *testing.T) {
	all := providers.GetAllCanonicalModels()
	if len(all) < 20 {
		t.Fatalf("expected at least 20 canonical models, got %d", len(all))
	}

	foundAntigravity := false
	foundClaude := false
	foundOpenAI := false
	foundGemini := false
	foundGroq := false

	for _, m := range all {
		switch m.ID {
		case "gemini-3.8-flash":
			foundAntigravity = true
		case "claude-3-7-sonnet-20250219":
			foundClaude = true
		case "gpt-4o":
			foundOpenAI = true
		case "gemini-2.5-flash":
			foundGemini = true
		case "llama-3.3-70b-versatile":
			foundGroq = true
		}
	}

	if !foundAntigravity || !foundClaude || !foundOpenAI || !foundGemini || !foundGroq {
		t.Errorf("missing core canonical models from 9router catalog")
	}

	h := &httpapi.GatewayHandler{}
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()

	h.ListModels(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var listResp struct {
		Object string `json:"object"`
		Data   []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode models list: %v", err)
	}

	if len(listResp.Data) < 20 {
		t.Errorf("expected at least 20 models in /v1/models response, got %d", len(listResp.Data))
	}
}

func Test9Router_MockGatewayExecution(t *testing.T) {
	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)
	registry := providers.NewRegistry()

	mockAdapter := &mockTestAdapter{}
	registry.Register("mock", mockAdapter)

	router.SetProvider("p-mock", "mock")
	router.SetAccount(&routing.Account{
		ID:         "acc-mock-1",
		ProviderID: "p-mock",
		State:      "active",
		Enabled:    true,
		Priority:   1,
	})

	resolver := func(ctx context.Context, accountID string) (*providers.Credentials, error) {
		return &providers.Credentials{APIKey: "mock-key"}, nil
	}

	gw := gateway.NewGateway(router, registry, nil, cd, nil, resolver)

	req := &providers.Request{
		Model: "mock/test-model",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello from 9Router test"},
		},
	}

	resp, err := gw.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("gw.Execute failed: %v", err)
	}
	if resp.Content != "mock response" {
		t.Errorf("unexpected content: %s", resp.Content)
	}
}

type mockTestAdapter struct{}

func (m *mockTestAdapter) Kind() string {
	return "mock"
}

func (m *mockTestAdapter) Models(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error) {
	return []providers.ModelInfo{{ID: "test-model", Name: "Test Model"}}, nil
}

func (m *mockTestAdapter) Execute(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
	return &providers.Response{
		ID:      "resp-1",
		Model:   req.Model,
		Content: "mock response",
		Usage: providers.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func (m *mockTestAdapter) ExecuteStream(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
	ch := make(chan providers.StreamEvent, 2)
	ch <- providers.StreamEvent{Type: providers.StreamEventDelta, Delta: "mock stream"}
	ch <- providers.StreamEvent{Type: providers.StreamEventDone}
	close(ch)
	return ch, nil
}
