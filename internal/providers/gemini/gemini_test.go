package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dresar/ekarouter/internal/providers"
)

func TestGeminiAdapterExecuteAndStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, ":generateContent") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"candidates": [
					{
						"content": {
							"role": "model",
							"parts": [
								{"text": "thinking step...", "thought": true},
								{"text": "Hello Gemini!"}
							]
						},
						"finishReason": "STOP"
					}
				],
				"usageMetadata": {
					"promptTokenCount": 10,
					"candidatesTokenCount": 8,
					"totalTokenCount": 18
				}
			}`))
			return
		}
		if strings.Contains(r.URL.Path, ":streamGenerateContent") {
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprintf(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"Hello \",\"thought\":false}]}}]}\n\n")
			fmt.Fprintf(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"Gemini!\",\"thought\":false}]}}],\"usageMetadata\":{\"promptTokenCount\":10,\"candidatesTokenCount\":8,\"totalTokenCount\":18}}\n\n")
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	creds := &providers.Credentials{BaseURL: server.URL, APIKey: "test-gemini-key"}
	req := &providers.Request{
		Model: "gemini-2.5-flash",
		Messages: []providers.Message{
			{Role: "user", Content: "Hi"},
		},
	}

	ctx := context.Background()
	res, err := adapter.Execute(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if res.Content != "Hello Gemini!" {
		t.Errorf("expected Hello Gemini!, got %s", res.Content)
	}
	if res.Reasoning != "thinking step..." {
		t.Errorf("expected thinking step..., got %s", res.Reasoning)
	}
	if res.Usage.TotalTokens != 18 {
		t.Errorf("expected 18 tokens, got %d", res.Usage.TotalTokens)
	}

	streamChan, err := adapter.ExecuteStream(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var streamContent string
	var gotUsage bool
	for ev := range streamChan {
		if ev.Type == providers.StreamEventDelta {
			streamContent += ev.Delta
		}
		if ev.Type == providers.StreamEventUsage {
			gotUsage = true
		}
	}
	if streamContent != "Hello Gemini!" {
		t.Errorf("expected Hello Gemini!, got %s", streamContent)
	}
	if !gotUsage {
		t.Error("expected usage event")
	}
}

func TestAntigravityAdapterWrapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "Antigravity-IDE/1.0" {
			t.Errorf("unexpected User-Agent: %s", r.Header.Get("User-Agent"))
		}
		var envelope map[string]any
		if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil {
			t.Fatalf("failed to decode envelope: %v", err)
		}
		if envelope["projectId"] != "my-project-id" {
			t.Errorf("expected my-project-id, got %v", envelope["projectId"])
		}
		if envelope["model"] != "gemini-2.5-pro" {
			t.Errorf("expected gemini-2.5-pro, got %v", envelope["model"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"Antigravity OK"}]},"finishReason":"STOP"}]}`))
	}))
	defer server.Close()

	adapter := NewAntigravityAdapter(server.Client())
	if adapter.Kind() != "antigravity" {
		t.Errorf("expected kind antigravity, got %s", adapter.Kind())
	}

	creds := &providers.Credentials{
		BaseURL:     server.URL,
		ProjectID:   "my-project-id",
		AccessToken: "ya29.test-token",
	}
	req := &providers.Request{
		Model: "gemini-2.5-pro",
		Messages: []providers.Message{
			{Role: "system", Content: "You are Antigravity."},
			{Role: "user", Content: "Hello"},
		},
	}

	res, err := adapter.Execute(context.Background(), req, creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "Antigravity OK" {
		t.Errorf("expected Antigravity OK, got %s", res.Content)
	}
}

func TestGeminiCLIAdapterWrapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "gemini-cli/1.0.0" {
			t.Errorf("unexpected User-Agent: %s", r.Header.Get("User-Agent"))
		}
		var envelope map[string]any
		if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil {
			t.Fatalf("failed to decode envelope: %v", err)
		}
		if envelope["project"] != "cli-project" {
			t.Errorf("expected cli-project, got %v", envelope["project"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"CLI OK"}]},"finishReason":"STOP"}]}`))
	}))
	defer server.Close()

	adapter := NewCLIAdapter(server.Client())
	if adapter.Kind() != "gemini-cli" {
		t.Errorf("expected kind gemini-cli, got %s", adapter.Kind())
	}

	creds := &providers.Credentials{
		BaseURL:     server.URL,
		ProjectID:   "cli-project",
		AccessToken: "test-cli-token",
	}
	req := &providers.Request{
		Model: "gemini-2.0-flash",
		Messages: []providers.Message{
			{Role: "user", Content: "Run command"},
		},
	}

	res, err := adapter.Execute(context.Background(), req, creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "CLI OK" {
		t.Errorf("expected CLI OK, got %s", res.Content)
	}
}

func TestToolSanitizationAndAspectRatios(t *testing.T) {
	sanitized := SanitizeToolName("run-bash.command@v1!")
	if strings.Contains(sanitized, "@") || strings.Contains(sanitized, "!") {
		t.Errorf("expected tool name to be sanitized, got %s", sanitized)
	}

	invalidStart := SanitizeToolName("123tool")
	if !strings.HasPrefix(invalidStart, "tool_") {
		t.Errorf("expected leading number to be prefixed with tool_, got %s", invalidStart)
	}

	ar1 := ExtractAspectRatio("generate an image with 16:9 ratio")
	if ar1 != "16:9" {
		t.Errorf("expected 16:9, got %s", ar1)
	}

	ar2 := ExtractAspectRatio("photo portrait in 9:16")
	if ar2 != "9:16" {
		t.Errorf("expected 9:16, got %s", ar2)
	}

	arDefault := ExtractAspectRatio("square photo")
	if arDefault != "1:1" {
		t.Errorf("expected 1:1, got %s", arDefault)
	}
}

func TestThoughtSignatureStore(t *testing.T) {
	store := NewThoughtSignatureStore()
	store.Store("sess-1", "signature-abc")
	if store.Get("sess-1") != "signature-abc" {
		t.Errorf("expected signature-abc, got %s", store.Get("sess-1"))
	}
	store.Clear("sess-1")
	if store.Get("sess-1") != "" {
		t.Errorf("expected empty string after clear, got %s", store.Get("sess-1"))
	}
}

func TestGeminiCapabilities(t *testing.T) {
	models := DefaultModels()
	if len(models) == 0 {
		t.Fatal("expected at least one model")
	}
	capsPro := ModelCapabilities("gemini-2.5-pro")
	if !capsPro.Reasoning || !capsPro.Vision || !capsPro.ToolCalling {
		t.Errorf("expected pro model to have reasoning, vision, and tool calling")
	}

	capsLite := ModelCapabilities("gemini-2.0-flash-lite")
	if capsLite.Reasoning {
		t.Errorf("expected lite model to have reasoning false")
	}
}
