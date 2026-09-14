package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/providers"
)

func TestOpenAIAdapterExecuteAndStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chat/completions" {
			if r.Header.Get("Accept") == "text/event-stream" {
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello \",\"reasoning_content\":\"thought-1 \"}}]}\n\n")
				fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"World!\"}}]}\n\n")
				fmt.Fprintf(w, "data: [DONE]\n\n")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":"chat-1","model":"gpt-4o","choices":[{"message":{"role":"assistant","content":"Hello!","reasoning_content":"thought-ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	creds := &providers.Credentials{BaseURL: server.URL, APIKey: "test-key"}
	req := &providers.Request{
		Model: "gpt-4o",
		Messages: []providers.Message{
			{Role: "user", Content: "Hi"},
		},
	}

	ctx := context.Background()
	res, err := adapter.Execute(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if res.Content != "Hello!" {
		t.Errorf("expected Hello!, got %s", res.Content)
	}
	if res.Reasoning != "thought-ok" {
		t.Errorf("expected thought-ok, got %s", res.Reasoning)
	}
	if res.Usage.TotalTokens != 15 {
		t.Errorf("expected 15 tokens, got %d", res.Usage.TotalTokens)
	}

	streamChan, err := adapter.ExecuteStream(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var streamContent string
	var streamReasoning string
	done := false
	for ev := range streamChan {
		if ev.Type == providers.StreamEventDelta {
			streamContent += ev.Delta
		}
		if ev.Type == providers.StreamEventReasoning {
			streamReasoning += ev.Reasoning
		}
		if ev.Type == providers.StreamEventDone {
			done = true
		}
	}
	if streamContent != "Hello World!" {
		t.Errorf("expected 'Hello World!', got '%s'", streamContent)
	}
	if streamReasoning != "thought-1 " {
		t.Errorf("expected 'thought-1 ', got '%s'", streamReasoning)
	}
	if !done {
		t.Error("expected stream done event")
	}
}

func TestReasoningModelRoleConversionAndEffort(t *testing.T) {
	adapter := NewAdapter(nil)
	maxTokens := 1000
	req := &providers.Request{
		Model: "o3-mini",
		Messages: []providers.Message{
			{Role: "system", Content: "System prompt"},
			{Role: "user", Content: "Solve this"},
		},
		ReasoningEffort: "high",
		MaxTokens:       &maxTokens,
	}

	body := adapter.BuildRequestBody(req, false)
	messages, ok := body["messages"].([]map[string]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("unexpected messages: %v", body["messages"])
	}
	if messages[0]["role"] != "developer" {
		t.Errorf("expected developer role for o3-mini system message, got %v", messages[0]["role"])
	}
	if body["reasoning_effort"] != "high" {
		t.Errorf("expected reasoning_effort high, got %v", body["reasoning_effort"])
	}
	if _, hasMax := body["max_tokens"]; hasMax {
		t.Error("expected max_tokens to be removed in favor of max_completion_tokens for reasoning model")
	}
	if body["max_completion_tokens"] != 1000 {
		t.Errorf("expected max_completion_tokens 1000, got %v", body["max_completion_tokens"])
	}
}

func TestStripServerIDs(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"rs_12345", "12345"},
		{"fc_abcde", "abcde"},
		{"resp_xyz", "xyz"},
		{"msg_999", "999"},
		{"normal_id", "normal_id"},
	}

	for _, c := range cases {
		got := StripServerID(c.input)
		if got != c.expected {
			t.Errorf("StripServerID(%s) = %s; want %s", c.input, got, c.expected)
		}
	}
}

func TestLargeSSELineExceedsDefaultBuffer(t *testing.T) {
	largeContent := strings.Repeat("A", 128*1024)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"%s\"}}]}\n\n", largeContent)
		fmt.Fprintf(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	creds := &providers.Credentials{BaseURL: server.URL}
	req := &providers.Request{Model: "test"}

	ctx := context.Background()
	streamChan, err := adapter.ExecuteStream(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var totalLen int
	for ev := range streamChan {
		if ev.Type == providers.StreamEventDelta {
			totalLen += len(ev.Delta)
		}
		if ev.Type == providers.StreamEventError {
			t.Fatalf("unexpected stream error for large line: %v", ev.Error)
		}
	}

	if totalLen != len(largeContent) {
		t.Errorf("expected %d bytes, got %d", len(largeContent), totalLen)
	}
}

func TestCancellationDoesNotLeak(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		for i := 0; i < 50; i++ {
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Chunk\"}}]}\n\n")
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	creds := &providers.Credentials{BaseURL: server.URL}
	req := &providers.Request{Model: "test"}

	ctx, cancel := context.WithCancel(context.Background())
	streamChan, err := adapter.ExecuteStream(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected stream start error: %v", err)
	}

	<-streamChan
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestOpenAICapabilities(t *testing.T) {
	capsO1 := ModelCapabilities("o1")
	if !capsO1.Reasoning {
		t.Error("expected o1 to have reasoning capability")
	}

	caps4o := ModelCapabilities("gpt-4o")
	if caps4o.Reasoning {
		t.Error("expected gpt-4o to have reasoning false")
	}
}

func TestCodexAdapterExecuteAndStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/responses" {
			if r.Header.Get("Accept") == "text/event-stream" {
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprintf(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"Codex stream\"}\n\n")
				fmt.Fprintf(w, "data: [DONE]\n\n")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"id":"resp_123",
				"model":"codex-mini",
				"status":"completed",
				"output":[
					{
						"type":"message",
						"role":"assistant",
						"content":[{"type":"output_text","text":"Codex OK"}]
					}
				],
				"usage":{"input_tokens":8,"output_tokens":4}
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := NewCodexAdapter(server.Client())
	if adapter.Kind() != "codex" {
		t.Errorf("expected codex kind, got %s", adapter.Kind())
	}

	creds := &providers.Credentials{BaseURL: server.URL, APIKey: "test-codex-key"}
	req := &providers.Request{
		Model: "codex-mini",
		Messages: []providers.Message{
			{Role: "system", Content: "Instructions"},
			{Role: "user", Content: "Hello"},
		},
	}

	res, err := adapter.Execute(context.Background(), req, creds)
	if err != nil {
		t.Fatalf("unexpected codex execute error: %v", err)
	}
	if res.Content != "Codex OK" {
		t.Errorf("expected Codex OK, got %s", res.Content)
	}
	if res.Usage.TotalTokens != 12 {
		t.Errorf("expected 12 total tokens, got %d", res.Usage.TotalTokens)
	}

	streamChan, err := adapter.ExecuteStream(context.Background(), req, creds)
	if err != nil {
		t.Fatalf("unexpected codex stream error: %v", err)
	}
	var streamText string
	for ev := range streamChan {
		if ev.Type == providers.StreamEventDelta {
			streamText += ev.Delta
		}
	}
	if streamText != "Codex stream" {
		t.Errorf("expected 'Codex stream', got '%s'", streamText)
	}
}

func TestCodexToolNormalization(t *testing.T) {
	tools := []any{
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "test_func",
				"description": "A test function",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type":    "string",
							"pattern": `^\p{L}+$`,
						},
					},
				},
			},
		},
	}

	normalized := NormalizeCodexTools(tools)
	if len(normalized) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(normalized))
	}
	if normalized[0]["name"] != "test_func" {
		t.Errorf("expected test_func, got %v", normalized[0]["name"])
	}
	params, ok := normalized[0]["parameters"].(map[string]any)
	if !ok {
		t.Fatalf("expected parameters map, got %v", normalized[0]["parameters"])
	}
	props, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties map, got %v", params["properties"])
	}
	nameProp, ok := props["name"].(map[string]any)
	if !ok {
		t.Fatalf("expected name prop, got %v", props["name"])
	}
	if pattern, ok := nameProp["pattern"].(string); ok && strings.Contains(pattern, `\p{`) {
		t.Errorf("expected unicode escape to be stripped, got %s", pattern)
	}
}
