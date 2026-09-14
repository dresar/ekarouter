package mocks

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"github.com/dresar/ekarouter/internal/providers"
)

type UpstreamBehavior struct {
	StatusCode      int
	ResponseBody    string
	ContentType     string
	Delay           time.Duration
	StreamChunks    []string
	Trigger429Count int32
}

type MockUpstreamServer struct {
	Server       *httptest.Server
	URL          string
	RequestCount int32
	Behavior     UpstreamBehavior
}

func NewMockUpstreamServer(b UpstreamBehavior) *MockUpstreamServer {
	m := &MockUpstreamServer{
		Behavior: b,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&m.RequestCount, 1)

		if m.Behavior.Delay > 0 {
			time.Sleep(m.Behavior.Delay)
		}

		if m.Behavior.Trigger429Count > 0 && atomic.LoadInt32(&m.RequestCount) <= m.Behavior.Trigger429Count {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"Rate limit exceeded","type":"rate_limit_error","code":"rate_limit_exceeded"}}`))
			return
		}

		if len(m.Behavior.StreamChunks) > 0 {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.WriteHeader(http.StatusOK)

			flusher, ok := w.(http.Flusher)
			for _, chunk := range m.Behavior.StreamChunks {
				fmt.Fprintf(w, "data: %s\n\n", chunk)
				if ok {
					flusher.Flush()
				}
				time.Sleep(5 * time.Millisecond)
			}
			fmt.Fprintf(w, "data: [DONE]\n\n")
			if ok {
				flusher.Flush()
			}
			return
		}

		ct := m.Behavior.ContentType
		if ct == "" {
			ct = "application/json"
		}
		w.Header().Set("Content-Type", ct)

		code := m.Behavior.StatusCode
		if code == 0 {
			code = http.StatusOK
		}
		w.WriteHeader(code)

		resp := m.Behavior.ResponseBody
		if resp == "" && code == http.StatusOK {
			resp = `{"id":"chatcmpl-mock-123","object":"chat.completion","created":1700000000,"model":"mock-model","choices":[{"index":0,"message":{"role":"assistant","content":"Hello from mock!"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`
		}
		_, _ = w.Write([]byte(resp))
	})

	m.Server = httptest.NewServer(mux)
	m.URL = m.Server.URL
	return m
}

func (m *MockUpstreamServer) Close() {
	if m.Server != nil {
		m.Server.Close()
	}
}

type MockAIAdapter struct {
	KindName  string
	ExecuteFn func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error)
	StreamFn  func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error)
	ModelsFn  func(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error)
	CallCount int32
}

func (m *MockAIAdapter) Kind() string {
	if m.KindName != "" {
		return m.KindName
	}
	return "mock-ai"
}

func (m *MockAIAdapter) Models(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error) {
	if m.ModelsFn != nil {
		return m.ModelsFn(ctx, creds)
	}
	return []providers.ModelInfo{
		{ID: "mock-model-1", Name: "Mock Model 1"},
		{ID: "mock-model-2", Name: "Mock Model 2"},
	}, nil
}

func (m *MockAIAdapter) Execute(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
	atomic.AddInt32(&m.CallCount, 1)
	if m.ExecuteFn != nil {
		return m.ExecuteFn(ctx, req, creds)
	}
	return &providers.Response{
		ID:           "mock-resp-" + req.Model,
		Model:        req.Model,
		Role:         "assistant",
		Content:      "Mock response from " + req.Model,
		FinishReason: "stop",
		Usage: providers.Usage{
			PromptTokens:     10,
			CompletionTokens: 12,
			TotalTokens:      22,
		},
	}, nil
}

func (m *MockAIAdapter) ExecuteStream(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
	atomic.AddInt32(&m.CallCount, 1)
	if m.StreamFn != nil {
		return m.StreamFn(ctx, req, creds)
	}

	ch := make(chan providers.StreamEvent, 3)
	go func() {
		ch <- providers.StreamEvent{Type: providers.StreamEventDelta, Delta: "Mock "}
		ch <- providers.StreamEvent{Type: providers.StreamEventDelta, Delta: "Stream!"}
		ch <- providers.StreamEvent{Type: providers.StreamEventDone}
		close(ch)
	}()
	return ch, nil
}
