package gateway

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
)

type mockAdapter struct {
	kind      string
	executeFn func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error)
	streamFn  func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error)
}

func (m *mockAdapter) Kind() string { return m.kind }
func (m *mockAdapter) Models(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error) {
	return nil, nil
}
func (m *mockAdapter) Execute(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
	return m.executeFn(ctx, req, creds)
}
func (m *mockAdapter) ExecuteStream(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
	return m.streamFn(ctx, req, creds)
}

func TestGatewayFallbackOnTransientError(t *testing.T) {
	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)

	router.SetAccount(&routing.Account{ID: "acc-1", ProviderID: "p1", State: "active", Enabled: true})
	router.SetAccount(&routing.Account{ID: "acc-2", ProviderID: "p2", State: "active", Enabled: true})

	route := &routing.Route{
		Name:     "gpt-4o",
		Strategy: routing.StrategyPriority,
		Enabled:  true,
		Items: []routing.RouteItem{
			{ProviderID: "p1", ProviderKind: "kind1", AccountID: "acc-1", ModelName: "gpt-4o", Priority: 10, Enabled: true, MaxRetries: 1},
			{ProviderID: "p2", ProviderKind: "kind2", AccountID: "acc-2", ModelName: "gpt-4o", Priority: 20, Enabled: true, MaxRetries: 1},
		},
	}
	router.SetRoute(route)

	reg := providers.NewRegistry()
	reg.Register("kind1", &mockAdapter{
		kind: "kind1",
		executeFn: func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
			return nil, providers.ClassifyHTTPError(http.StatusTooManyRequests, "rate limit")
		},
	})
	reg.Register("kind2", &mockAdapter{
		kind: "kind2",
		executeFn: func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
			return &providers.Response{
				ID:      "resp-fallback",
				Content: "Hello from fallback provider!",
			}, nil
		},
	})

	credResolver := func(ctx context.Context, accountID string) (*providers.Credentials, error) {
		return &providers.Credentials{APIKey: "key-" + accountID}, nil
	}

	gw := NewGateway(router, reg, tokensaver.New("off"), cd, nil, credResolver)

	resp, err := gw.Execute(context.Background(), &providers.Request{
		ID:    "req-test",
		Model: "gpt-4o",
	})
	if err != nil {
		t.Fatalf("expected successful fallback, got error: %v", err)
	}
	if resp.Content != "Hello from fallback provider!" {
		t.Errorf("expected fallback content, got %s", resp.Content)
	}
	if !cd.IsCoolingDown("acc-1") {
		t.Error("expected acc-1 to be in cooldown after 429 failure")
	}
}

func TestGatewayNoFallbackOnClientError(t *testing.T) {
	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)

	router.SetAccount(&routing.Account{ID: "acc-1", ProviderID: "p1", State: "active", Enabled: true})
	router.SetAccount(&routing.Account{ID: "acc-2", ProviderID: "p2", State: "active", Enabled: true})

	route := &routing.Route{
		Name:     "model-test",
		Strategy: routing.StrategyPriority,
		Enabled:  true,
		Items: []routing.RouteItem{
			{ProviderID: "p1", ProviderKind: "kind1", AccountID: "acc-1", ModelName: "model-test", Priority: 10, Enabled: true},
			{ProviderID: "p2", ProviderKind: "kind2", AccountID: "acc-2", ModelName: "model-test", Priority: 20, Enabled: true},
		},
	}
	router.SetRoute(route)

	reg := providers.NewRegistry()
	var kind2Called bool
	reg.Register("kind1", &mockAdapter{
		kind: "kind1",
		executeFn: func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
			return nil, providers.ClassifyHTTPError(http.StatusBadRequest, "invalid parameter")
		},
	})
	reg.Register("kind2", &mockAdapter{
		kind: "kind2",
		executeFn: func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
			kind2Called = true
			return &providers.Response{Content: "Should not be reached"}, nil
		},
	})

	credResolver := func(ctx context.Context, accountID string) (*providers.Credentials, error) {
		return &providers.Credentials{APIKey: "key"}, nil
	}

	gw := NewGateway(router, reg, tokensaver.New("off"), cd, nil, credResolver)

	_, err := gw.Execute(context.Background(), &providers.Request{
		ID:    "req-client-err",
		Model: "model-test",
	})
	if err == nil {
		t.Fatal("expected client error, got nil")
	}
	if kind2Called {
		t.Error("expected kind2 to NOT be called for non-retryable 400 error")
	}
}

func TestGatewayExecuteStream(t *testing.T) {
	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)

	router.SetAccount(&routing.Account{ID: "acc-1", ProviderID: "p1", State: "active", Enabled: true})
	route := &routing.Route{
		Name:     "stream-model",
		Strategy: routing.StrategyPriority,
		Enabled:  true,
		Items: []routing.RouteItem{
			{ProviderID: "p1", ProviderKind: "kind1", AccountID: "acc-1", ModelName: "stream-model", Priority: 10, Enabled: true},
		},
	}
	router.SetRoute(route)

	reg := providers.NewRegistry()
	reg.Register("kind1", &mockAdapter{
		kind: "kind1",
		streamFn: func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
			ch := make(chan providers.StreamEvent, 2)
			ch <- providers.StreamEvent{Type: providers.StreamEventDelta, Delta: "Part 1"}
			ch <- providers.StreamEvent{Type: providers.StreamEventDone}
			close(ch)
			return ch, nil
		},
	})

	credResolver := func(ctx context.Context, accountID string) (*providers.Credentials, error) {
		return &providers.Credentials{APIKey: "key"}, nil
	}

	gw := NewGateway(router, reg, tokensaver.New("safe"), cd, nil, credResolver)

	streamChan, err := gw.ExecuteStream(context.Background(), &providers.Request{
		ID:    "req-stream",
		Model: "stream-model",
	})
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var chunks []string
	for ev := range streamChan {
		if ev.Type == providers.StreamEventDelta {
			chunks = append(chunks, ev.Delta)
		}
	}

	if len(chunks) != 1 || chunks[0] != "Part 1" {
		t.Errorf("unexpected stream chunks: %v", chunks)
	}
}

func TestGatewayStreamClientCancellation(t *testing.T) {
	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)
	router.SetAccount(&routing.Account{ID: "acc-1", ProviderID: "p1", State: "active", Enabled: true})
	router.SetRoute(&routing.Route{
		Name:     "cancel-model",
		Strategy: routing.StrategyPriority,
		Enabled:  true,
		Items: []routing.RouteItem{
			{ProviderID: "p1", ProviderKind: "kind1", AccountID: "acc-1", ModelName: "cancel-model", Priority: 10, Enabled: true},
		},
	})

	reg := providers.NewRegistry()
	reg.Register("kind1", &mockAdapter{
		kind: "kind1",
		streamFn: func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
			ch := make(chan providers.StreamEvent, 50)
			for i := 0; i < 20; i++ {
				ch <- providers.StreamEvent{Type: providers.StreamEventDelta, Delta: "chunk"}
			}
			return ch, nil
		},
	})

	credResolver := func(ctx context.Context, accountID string) (*providers.Credentials, error) {
		return &providers.Credentials{APIKey: "key"}, nil
	}

	gw := NewGateway(router, reg, tokensaver.New("safe"), cd, nil, credResolver)
	ctx, cancel := context.WithCancel(context.Background())

	streamChan, err := gw.ExecuteStream(ctx, &providers.Request{
		ID:    "req-cancel",
		Model: "cancel-model",
	})
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	<-streamChan
	cancel()
}

func TestGatewayTokenSaverOptOut(t *testing.T) {
	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)
	gw := NewGateway(router, nil, tokensaver.New("safe"), cd, nil, nil)

	diffText := "diff --git a/file b/file\n" + strings.Repeat("+ added line\n", 50)
	reqOptOut := &providers.Request{
		OptOutTokenSaver: true,
		Messages: []providers.Message{
			{Role: "user", Content: diffText},
		},
	}
	prep := gw.PrepareRequest(reqOptOut)
	if prep.Messages[0].Content != diffText {
		t.Errorf("expected untouched text when OptOutTokenSaver is true")
	}
}

func TestGatewayFallbackOnAuthError(t *testing.T) {
	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)

	router.SetAccount(&routing.Account{ID: "acc-expired", ProviderID: "p1", State: "active", Enabled: true})
	router.SetAccount(&routing.Account{ID: "acc-healthy", ProviderID: "p2", State: "active", Enabled: true})

	route := &routing.Route{
		Name:     "gpt-4o",
		Strategy: routing.StrategyPriority,
		Enabled:  true,
		Items: []routing.RouteItem{
			{ProviderID: "p1", ProviderKind: "kind1", AccountID: "acc-expired", ModelName: "gpt-4o", Priority: 10, Enabled: true, MaxRetries: 1},
			{ProviderID: "p2", ProviderKind: "kind2", AccountID: "acc-healthy", ModelName: "gpt-4o", Priority: 20, Enabled: true, MaxRetries: 1},
		},
	}
	router.SetRoute(route)

	reg := providers.NewRegistry()
	reg.Register("kind1", &mockAdapter{
		kind: "kind1",
		executeFn: func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
			return nil, providers.ClassifyHTTPError(http.StatusUnauthorized, "invalid api key")
		},
	})
	reg.Register("kind2", &mockAdapter{
		kind: "kind2",
		executeFn: func(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
			return &providers.Response{
				ID:      "resp-auth-fallback",
				Model:   "gpt-4o",
				Role:    "assistant",
				Content: "fallback success after 401",
			}, nil
		},
	})

	credResolver := func(ctx context.Context, accountID string) (*providers.Credentials, error) {
		return &providers.Credentials{APIKey: "key-" + accountID}, nil
	}

	gw := NewGateway(router, reg, nil, cd, nil, credResolver)
	resp, err := gw.Execute(context.Background(), &providers.Request{
		Model: "gpt-4o",
		Messages: []providers.Message{
			{Role: "user", Content: "hello"},
		},
	})

	if err != nil {
		t.Fatalf("expected successful failover, got error: %v", err)
	}
	if resp.Content != "fallback success after 401" {
		t.Errorf("unexpected content: %s", resp.Content)
	}
	if !cd.IsCoolingDown("acc-expired") {
		t.Errorf("expected acc-expired to be put into cooldown")
	}
}
