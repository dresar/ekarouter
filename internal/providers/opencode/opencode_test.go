package opencode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dresar/ekarouter/internal/providers"
)

func TestOpenCodeAdapter(t *testing.T) {
	var capturedHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get(ClientHeader)
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"id": "chatcmpl-oc-1",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "Hello from OpenCode Free!",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	if adapter.Name() != "opencode" {
		t.Fatalf("expected name opencode, got %s", adapter.Name())
	}

	creds := &providers.Credentials{BaseURL: server.URL}
	req := &providers.Request{
		Model: "muse-spark-1.2-contributor-free",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	res, err := adapter.Execute(context.Background(), req, creds)
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if res.Content != "Hello from OpenCode Free!" {
		t.Errorf("expected Hello from OpenCode Free!, got %s", res.Content)
	}
	if capturedHeader != ClientValue {
		t.Errorf("expected %s=%s, got %s", ClientHeader, ClientValue, capturedHeader)
	}
}
