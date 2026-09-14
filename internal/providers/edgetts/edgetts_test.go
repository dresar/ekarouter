package edgetts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dresar/ekarouter/internal/providers"
)

func TestEdgeTTSAdapter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"id": "chatcmpl-et-1",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "Hello from Edge-TTS!",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	if adapter.Name() != "edge-tts" {
		t.Fatalf("expected name edge-tts, got %s", adapter.Name())
	}

	creds := &providers.Credentials{BaseURL: server.URL}
	req := &providers.Request{
		Model: "edge-tts-voice",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	res, err := adapter.Execute(context.Background(), req, creds)
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if res.Content != "Hello from Edge-TTS!" {
		t.Errorf("expected Hello from Edge-TTS!, got %s", res.Content)
	}
}
