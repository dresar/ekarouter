package providers

import (
	"context"
	"errors"
	"testing"
)

type mockAdapter struct {
	kind string
}

func (m *mockAdapter) Kind() string {
	return m.kind
}

func (m *mockAdapter) Models(ctx context.Context, creds *Credentials) ([]ModelInfo, error) {
	return []ModelInfo{{ID: "mock-model", Name: "Mock Model"}}, nil
}

func (m *mockAdapter) Execute(ctx context.Context, req *Request, creds *Credentials) (*Response, error) {
	return &Response{Model: req.Model, Content: "mock response"}, nil
}

func (m *mockAdapter) ExecuteStream(ctx context.Context, req *Request, creds *Credentials) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 1)
	ch <- StreamEvent{Type: StreamEventDelta, Delta: "mock delta"}
	close(ch)
	return ch, nil
}

func TestProviderRegistry(t *testing.T) {
	r := NewRegistry()
	mock := &mockAdapter{kind: "mock"}
	r.Register("mock", mock)

	adapter, err := r.Get("mock")
	if err != nil {
		t.Fatalf("expected adapter, got error: %v", err)
	}
	if adapter.Kind() != "mock" {
		t.Errorf("expected mock kind, got %s", adapter.Kind())
	}

	if !r.Has("mock") {
		t.Error("expected Has('mock') to be true")
	}

	if r.Has("nonexistent") {
		t.Error("expected Has('nonexistent') to be false")
	}

	list := r.List()
	if len(list) != 1 || list[0] != "mock" {
		t.Errorf("expected ['mock'], got %v", list)
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

	errQuota := ClassifyHTTPError(403, "resource_exhausted by quota limit")
	if errQuota.Class != ErrorClassQuota {
		t.Errorf("expected ErrorClassQuota, got %s", errQuota.Class)
	}
	if !errQuota.IsTransient() {
		t.Error("expected quota to be transient")
	}

	errOverloaded := ClassifyHTTPError(529, "server overloaded")
	if errOverloaded.Class != ErrorClassUpstream5xx {
		t.Errorf("expected ErrorClassUpstream5xx, got %s", errOverloaded.Class)
	}

	wrapped := &ProviderError{
		StatusCode: 500,
		Class:      ErrorClassUpstream5xx,
		Message:    "internal error",
		Err:        errors.New("sub-error"),
	}
	if wrapped.Unwrap() == nil {
		t.Error("expected sub-error unwrapping")
	}
	if wrapped.Error() == "" {
		t.Error("expected non-empty error string")
	}

	if IsTransient(nil) {
		t.Error("expected IsTransient(nil) to be false")
	}
	if IsTransient(errors.New("generic error")) {
		t.Error("expected IsTransient(generic) to be false")
	}
}
