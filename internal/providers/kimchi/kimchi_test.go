package kimchi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dresar/ekarouter/internal/providers"
)

func TestKimchiAdapter(t *testing.T) {
	var capturedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUA = r.Header.Get(UserAgentHeader)
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"id": "chatcmpl-kc-1",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "Hello from Kimchi Dev Free!",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	if adapter.Name() != "kimchi" {
		t.Fatalf("expected name kimchi, got %s", adapter.Name())
	}

	creds := &providers.Credentials{BaseURL: server.URL, APIKey: "kc_test"}
	req := &providers.Request{
		Model: "minimax-m3",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	res, err := adapter.Execute(context.Background(), req, creds)
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if res.Content != "Hello from Kimchi Dev Free!" {
		t.Errorf("expected Hello from Kimchi Dev Free!, got %s", res.Content)
	}
	if capturedUA != UserAgentValue {
		t.Errorf("expected User-Agent %s, got %s", UserAgentValue, capturedUA)
	}
}
