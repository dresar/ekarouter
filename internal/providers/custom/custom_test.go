package custom

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dresar/ekarouter/internal/providers"
)

func TestCustomAdapterExecuteAndStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chat/completions" {
			if r.Header.Get("Accept") == "text/event-stream" {
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Custom stream\"}}]}\n\n")
				fmt.Fprintf(w, "data: [DONE]\n\n")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":"custom-1","model":"llama3","choices":[{"message":{"role":"assistant","content":"Custom OK"},"finish_reason":"stop"}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	if adapter.Kind() != "custom" {
		t.Errorf("expected custom kind, got %s", adapter.Kind())
	}

	creds := &providers.Credentials{BaseURL: server.URL}
	req := &providers.Request{
		Model: "llama3",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello custom"},
		},
	}

	ctx := context.Background()
	res, err := adapter.Execute(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "Custom OK" {
		t.Errorf("expected Custom OK, got %s", res.Content)
	}

	streamChan, err := adapter.ExecuteStream(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var content string
	for ev := range streamChan {
		if ev.Type == providers.StreamEventDelta {
			content += ev.Delta
		}
	}
	if content != "Custom stream" {
		t.Errorf("expected 'Custom stream', got %s", content)
	}
}

func TestBackendAdapterKindsAndURLs(t *testing.T) {
	cases := []struct {
		backend     string
		expectedURL string
	}{
		{"ollama", "http://localhost:11434/v1"},
		{"vllm", "http://localhost:8000/v1"},
		{"deepseek", "https://api.deepseek.com/v1"},
		{"groq", "https://api.groq.com/openai/v1"},
		{"openrouter", "https://openrouter.ai/api/v1"},
	}

	for _, c := range cases {
		a := NewBackendAdapter(c.backend, nil)
		if a.Kind() != c.backend {
			t.Errorf("expected kind %s, got %s", c.backend, a.Kind())
		}
		resolved := ResolveBaseURL(c.backend, nil)
		if resolved != c.expectedURL {
			t.Errorf("ResolveBaseURL(%s) = %s; want %s", c.backend, resolved, c.expectedURL)
		}
	}
}

func TestBackendCapabilities(t *testing.T) {
	capsDS := BackendCapabilities(BackendDeepSeek)
	if !capsDS.Reasoning {
		t.Error("expected DeepSeek to have reasoning capability")
	}

	capsOllama := BackendCapabilities(BackendOllama)
	if capsOllama.PromptCaching {
		t.Error("expected Ollama to have prompt caching false")
	}
}

func TestOpenRouterCustomHeadersInjection(t *testing.T) {
	var gotReferer, gotTitle string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReferer = r.Header.Get("HTTP-Referer")
		gotTitle = r.Header.Get("X-Title")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"or-1","choices":[{"message":{"role":"assistant","content":"OR OK"}}]}`))
	}))
	defer server.Close()

	adapter := NewBackendAdapter("openrouter", server.Client())
	creds := &providers.Credentials{BaseURL: server.URL, APIKey: "test-or-key"}
	req := &providers.Request{
		Model: "anthropic/claude-3-opus",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello OpenRouter"},
		},
	}

	res, err := adapter.Execute(context.Background(), req, creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "OR OK" {
		t.Errorf("expected OR OK, got %s", res.Content)
	}
	if gotReferer != "https://github.com/dresar/ekarouter" {
		t.Errorf("expected HTTP-Referer header, got '%s'", gotReferer)
	}
	if gotTitle != "EkaRouter" {
		t.Errorf("expected X-Title header, got '%s'", gotTitle)
	}
}
