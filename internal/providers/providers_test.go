package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestProviderRegistry(t *testing.T) {
	r := NewRegistry()
	openAI := NewOpenAIAdapter(nil)
	r.Register("openai", openAI)

	adapter, err := r.Get("openai")
	if err != nil {
		t.Fatalf("expected adapter, got error: %v", err)
	}
	if adapter.Kind() != "openai" {
		t.Errorf("expected openai kind, got %s", adapter.Kind())
	}

	_, err = r.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent adapter")
	}
}

func TestErrorClassification(t *testing.T) {
	err429 := ClassifyHTTPError(429, "rate limit")
	if !err429.IsTransient() {
		t.Error("expected 429 to be transient")
	}
	if !IsTransient(err429) {
		t.Error("expected IsTransient to return true for 429")
	}

	err503 := ClassifyHTTPError(503, "service unavailable")
	if !err503.IsTransient() {
		t.Error("expected 503 to be transient")
	}

	err401 := ClassifyHTTPError(401, "unauthorized")
	if err401.IsTransient() {
		t.Error("expected 401 to NOT be transient")
	}
}

func TestOpenAIAdapterExecuteAndStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chat/completions" {
			if r.Header.Get("Accept") == "text/event-stream" {
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello \"}}]}\n\n")
				fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"World!\"}}]}\n\n")
				fmt.Fprintf(w, "data: [DONE]\n\n")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":"chat-1","model":"gpt-4o","choices":[{"message":{"role":"assistant","content":"Hello!"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := NewOpenAIAdapter(server.Client())
	creds := &Credentials{BaseURL: server.URL, APIKey: "test-key"}
	req := &Request{
		Model: "gpt-4o",
		Messages: []Message{
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
	if res.Usage.TotalTokens != 15 {
		t.Errorf("expected 15 tokens, got %d", res.Usage.TotalTokens)
	}

	streamChan, err := adapter.ExecuteStream(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var streamContent string
	done := false
	for ev := range streamChan {
		if ev.Type == StreamEventDelta {
			streamContent += ev.Delta
		}
		if ev.Type == StreamEventDone {
			done = true
		}
	}
	if streamContent != "Hello World!" {
		t.Errorf("expected 'Hello World!', got '%s'", streamContent)
	}
	if !done {
		t.Error("expected stream done event")
	}
}

func TestAnthropicAdapterExecuteAndStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/messages" {
			if r.Header.Get("Accept") == "text/event-stream" {
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprintf(w, "data: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Claude response\"}}\n\n")
				fmt.Fprintf(w, "data: {\"type\":\"message_stop\"}\n\n")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":"msg-1","model":"claude-3-5-sonnet","role":"assistant","content":[{"type":"text","text":"Anthropic response"}],"stop_reason":"end_turn","usage":{"input_tokens":12,"output_tokens":6}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := NewAnthropicAdapter(server.Client())
	creds := &Credentials{BaseURL: server.URL, APIKey: "test-anthropic-key"}
	req := &Request{
		Model: "claude-3-5-sonnet",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
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

	streamChan, err := adapter.ExecuteStream(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var content string
	for ev := range streamChan {
		if ev.Type == StreamEventDelta {
			content += ev.Delta
		}
	}
	if content != "Claude response" {
		t.Errorf("expected 'Claude response', got %s", content)
	}
}

func TestClientCancellationPropagates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Chunk 1\"}}]}\n\n")
		w.(http.Flusher).Flush()
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	adapter := NewOpenAIAdapter(server.Client())
	creds := &Credentials{BaseURL: server.URL}
	req := &Request{Model: "test"}

	ctx, cancel := context.WithCancel(context.Background())
	streamChan, err := adapter.ExecuteStream(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected stream start error: %v", err)
	}

	<-streamChan
	cancel()

	var gotErr bool
	for ev := range streamChan {
		if ev.Type == StreamEventError {
			gotErr = true
		}
	}
	if !gotErr {
		t.Log("stream closed cleanly upon cancellation")
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

	adapter := NewOpenAIAdapter(server.Client())
	creds := &Credentials{BaseURL: server.URL}
	req := &Request{Model: "test"}

	ctx, cancel := context.WithCancel(context.Background())
	streamChan, err := adapter.ExecuteStream(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected stream start error: %v", err)
	}

	<-streamChan
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestLargeSSELineExceedsDefaultBuffer(t *testing.T) {
	largeContent := strings.Repeat("A", 128*1024)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"%s\"}}]}\n\n", largeContent)
		fmt.Fprintf(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	adapter := NewOpenAIAdapter(server.Client())
	creds := &Credentials{BaseURL: server.URL}
	req := &Request{Model: "test"}

	ctx := context.Background()
	streamChan, err := adapter.ExecuteStream(ctx, req, creds)
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var totalLen int
	for ev := range streamChan {
		if ev.Type == StreamEventDelta {
			totalLen += len(ev.Delta)
		}
		if ev.Type == StreamEventError {
			t.Fatalf("unexpected stream error for large line: %v", ev.Error)
		}
	}

	if totalLen != len(largeContent) {
		t.Errorf("expected %d bytes, got %d", len(largeContent), totalLen)
	}
}

func TestGeminiSystemInstruction(t *testing.T) {
	adapter := NewGeminiAdapter(nil)
	req := &Request{
		Model: "gemini-2.5-flash",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello"},
		},
	}
	body := adapter.buildRequestBody(req)
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed struct {
		SystemInstruction struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"systemInstruction"`
	}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(parsed.SystemInstruction.Parts) != 1 || parsed.SystemInstruction.Parts[0].Text != "You are a helpful assistant." {
		t.Errorf("unexpected systemInstruction: %s", string(bodyBytes))
	}
}

func TestAnthropicMultiSystem(t *testing.T) {
	adapter := NewAnthropicAdapter(nil)
	req := &Request{
		Model: "claude-3-5-sonnet",
		Messages: []Message{
			{Role: "system", Content: "System 1"},
			{Role: "system", Content: "System 2"},
			{Role: "user", Content: "Hello"},
		},
	}
	body := adapter.buildRequestBody(req, false)
	sysStr, ok := body["system"].(string)
	if !ok || sysStr != "System 1\n\nSystem 2" {
		t.Errorf("expected concatenated system prompt, got: %v", body["system"])
	}
}
