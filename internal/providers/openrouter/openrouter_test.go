package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dresar/ekarouter/internal/providers"
)

func TestOpenRouterAdapter(t *testing.T) {
	var gotReferer, gotTitle string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReferer = r.Header.Get(RefererHeader)
		gotTitle = r.Header.Get(TitleHeader)
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"id": "chatcmpl-or-1",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "Hello from OpenRouter!",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	if adapter.Name() != "openrouter" {
		t.Fatalf("expected name openrouter, got %s", adapter.Name())
	}

	creds := &providers.Credentials{BaseURL: server.URL, APIKey: "sk-or-test"}
	req := &providers.Request{
		Model: "deepseek/deepseek-chat:free",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	res, err := adapter.Execute(context.Background(), req, creds)
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if res.Content != "Hello from OpenRouter!" {
		t.Errorf("expected Hello from OpenRouter!, got %s", res.Content)
	}
	if gotReferer != RefererValue {
		t.Errorf("expected %s, got %s", RefererValue, gotReferer)
	}
	if gotTitle != TitleValue {
		t.Errorf("expected %s, got %s", TitleValue, gotTitle)
	}
}
