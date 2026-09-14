package kilogateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dresar/ekarouter/internal/providers"
)

func TestKiloGatewayAdapter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"id": "chatcmpl-kg-1",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "Hello from Kilo Gateway Free!",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	if adapter.Name() != "kilo-gateway" {
		t.Fatalf("expected name kilo-gateway, got %s", adapter.Name())
	}

	creds := &providers.Credentials{BaseURL: server.URL, APIKey: "kgw_test"}
	req := &providers.Request{
		Model: "kilo-auto/free",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	res, err := adapter.Execute(context.Background(), req, creds)
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if res.Content != "Hello from Kilo Gateway Free!" {
		t.Errorf("expected Hello from Kilo Gateway Free!, got %s", res.Content)
	}
}
