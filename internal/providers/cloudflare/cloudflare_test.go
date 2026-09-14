package cloudflare

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dresar/ekarouter/internal/providers"
)

func TestCloudflareAdapter(t *testing.T) {
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"id": "chatcmpl-cf-1",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "Hello from Cloudflare Workers AI!",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	if adapter.Name() != "cloudflare-ai" {
		t.Fatalf("expected name cloudflare-ai, got %s", adapter.Name())
	}

	creds := &providers.Credentials{
		BaseURL:   server.URL + "/accounts/{accountId}/ai/v1",
		ProjectID: "cf_acc_123",
		APIKey:    "cf_token_test",
	}
	req := &providers.Request{
		Model: "@cf/meta/llama-3.3-70b-instruct-fp8-fast",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	res, err := adapter.Execute(context.Background(), req, creds)
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if res.Content != "Hello from Cloudflare Workers AI!" {
		t.Errorf("expected Hello from Cloudflare Workers AI!, got %s", res.Content)
	}
	if !strings.Contains(requestedPath, "cf_acc_123") {
		t.Errorf("expected path to contain account ID cf_acc_123, got %s", requestedPath)
	}
}
