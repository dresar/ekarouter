package anthropic

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dresar/ekarouter/internal/providers"
)

func TestAnthropicAdapterExecuteAndStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/messages" {
			if r.Header.Get("Accept") == "text/event-stream" {
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprintf(w, "data: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"thinking_delta\",\"thinking\":\"thinking step...\"}}\n\n")
				fmt.Fprintf(w, "data: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Claude response\"}}\n\n")
				fmt.Fprintf(w, "data: {\"type\":\"message_stop\"}\n\n")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"id":"msg-1",
				"model":"claude-3-7-sonnet-20250219",
				"role":"assistant",
				"content":[
					{"type":"thinking","thinking":"deep thoughts"},
					{"type":"text","text":"Anthropic response"}
				],
				"stop_reason":"end_turn",
				"usage":{"input_tokens":12,"output_tokens":6}
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	creds := &providers.Credentials{BaseURL: server.URL, APIKey: "test-anthropic-key"}
	req := &providers.Request{
		Model: "claude-3-7-sonnet-20250219",
		Messages: []providers.Message{
			{Role: "system", Content: "System prompt"},
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	res, err := adapter.Execute(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected anthropic error: %v", err)
	}
	if res.Content != "Anthropic response" {
		t.Errorf("expected 'Anthropic response', got %s", res.Content)
	}
	if res.Reasoning != "deep thoughts" {
		t.Errorf("expected 'deep thoughts', got %s", res.Reasoning)
	}
	if res.Usage.TotalTokens != 18 {
		t.Errorf("expected 18 tokens, got %d", res.Usage.TotalTokens)
	}

	streamChan, err := adapter.ExecuteStream(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var content string
	var reasoning string
	for ev := range streamChan {
		if ev.Type == providers.StreamEventDelta {
			content += ev.Delta
		}
		if ev.Type == providers.StreamEventReasoning {
			reasoning += ev.Reasoning
		}
	}
	if content != "Claude response" {
		t.Errorf("expected 'Claude response', got %s", content)
	}
	if reasoning != "thinking step..." {
		t.Errorf("expected 'thinking step...', got %s", reasoning)
	}
}

func TestThinkingConfiguration(t *testing.T) {
	adapter := NewAdapter(nil)
	budget := 2048
	req := &providers.Request{
		Model: "claude-3-7-sonnet-20250219",
		Messages: []providers.Message{
			{Role: "user", Content: "Explain relativity"},
		},
		ThinkingBudget: &budget,
	}

	body, hasThinking := adapter.BuildRequestBody(req, false)
	if !hasThinking {
		t.Fatal("expected hasThinking to be true")
	}
	thinkingMap, ok := body["thinking"].(map[string]any)
	if !ok || thinkingMap["type"] != "enabled" || thinkingMap["budget_tokens"] != 2048 {
		t.Errorf("unexpected thinking configuration: %v", body["thinking"])
	}
	if body["temperature"] != 1.0 {
		t.Errorf("expected temperature 1.0 when thinking is enabled, got %v", body["temperature"])
	}
}

func TestAnthropicMultiSystem(t *testing.T) {
	adapter := NewAdapter(nil)
	req := &providers.Request{
		Model: "claude-3-5-sonnet",
		Messages: []providers.Message{
			{Role: "system", Content: "System 1"},
			{Role: "system", Content: "System 2"},
			{Role: "user", Content: "Hello"},
		},
	}
	body, _ := adapter.BuildRequestBody(req, false)
	sysStr, ok := body["system"].(string)
	if !ok || sysStr != "System 1\n\nSystem 2" {
		t.Errorf("expected concatenated system prompt, got: %v", body["system"])
	}
}

func TestToolCloaking(t *testing.T) {
	cloaked := CloakToolName("bash")
	if cloaked != "bash_ide" {
		t.Errorf("expected bash_ide, got %s", cloaked)
	}
	doubleCloaked := CloakToolName("bash_ide")
	if doubleCloaked != "bash_ide" {
		t.Errorf("expected idempotent cloaking, got %s", doubleCloaked)
	}
	decloaked := DecloakToolName("bash_ide")
	if decloaked != "bash" {
		t.Errorf("expected bash, got %s", decloaked)
	}
}

func TestAnthropicBetaHeaders(t *testing.T) {
	h := BuildBetaHeaders(true)
	if !strings.Contains(h, "thinking-2024-11-20") {
		t.Errorf("expected thinking beta header, got %s", h)
	}
	if !strings.Contains(h, "prompt-caching-2024-07-31") {
		t.Errorf("expected prompt caching beta header, got %s", h)
	}
}
