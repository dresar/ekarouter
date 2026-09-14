package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
	"github.com/dresar/ekarouter/internal/usage"
)

type CredentialResolver func(ctx context.Context, accountID string) (*providers.Credentials, error)

type Gateway struct {
	router      *routing.Router
	registry    *providers.Registry
	tokensaver  *tokensaver.TokenSaver
	cooldowns   *routing.CooldownManager
	recorder    *usage.Recorder
	credResolve CredentialResolver
}

func NewGateway(
	router *routing.Router,
	registry *providers.Registry,
	ts *tokensaver.TokenSaver,
	cd *routing.CooldownManager,
	rec *usage.Recorder,
	credResolve CredentialResolver,
) *Gateway {
	if cd == nil {
		cd = routing.NewCooldownManager()
	}
	return &Gateway{
		router:      router,
		registry:    registry,
		tokensaver:  ts,
		cooldowns:   cd,
		recorder:    rec,
		credResolve: credResolve,
	}
}

func (g *Gateway) PrepareRequest(req *providers.Request) *providers.Request {
	if req.OptOutTokenSaver || g.tokensaver == nil || g.tokensaver.Mode() == tokensaver.ModeOff {
		return req
	}

	for i := range req.Messages {
		if req.Messages[i].Role == "user" || req.Messages[i].Role == "tool" {
			req.Messages[i].Content = g.tokensaver.Compact(req.Messages[i].Content)
		}
	}
	return req
}

func (g *Gateway) Execute(ctx context.Context, req *providers.Request) (*providers.Response, error) {
	start := time.Now()
	prepared := g.PrepareRequest(req)

	targets, err := g.router.SelectTargets(prepared.Model)
	if err != nil {
		return nil, fmt.Errorf("route selection: %w", err)
	}

	var lastErr error

	for _, target := range targets {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		adapter, err := g.registry.Get(target.ProviderKind)
		if err != nil {
			lastErr = err
			continue
		}

		creds, err := g.credResolve(ctx, target.AccountID)
		if err != nil {
			lastErr = err
			continue
		}

		targetReq := *prepared
		targetReq.Model = target.ModelName

		attempts := target.MaxRetries
		if attempts <= 0 {
			attempts = 1
		}

		for attempt := 1; attempt <= attempts; attempt++ {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			reqCtx, cancel := context.WithTimeout(ctx, target.Timeout)
			resp, execErr := adapter.Execute(reqCtx, &targetReq, creds)
			cancel()

			if execErr == nil {
				g.cooldowns.MarkSuccess(target.AccountID)
				g.recordUsage(req.ID, target, resp.Usage, time.Since(start), 200, "")
				return resp, nil
			}

			lastErr = execErr

			if !providers.IsTransient(execErr) {
				g.recordUsage(req.ID, target, providers.Usage{}, time.Since(start), 400, execErr.Error())
				return nil, execErr
			}

			if attempt < attempts {
				time.Sleep(time.Duration(attempt*50) * time.Millisecond)
			}
		}

		g.cooldowns.MarkFailure(target.AccountID, 15*time.Second)
	}

	g.recordUsage(req.ID, routing.Target{}, providers.Usage{}, time.Since(start), 503, "all routes failed")
	if lastErr != nil {
		return nil, fmt.Errorf("all routes exhausted, last error: %w", lastErr)
	}
	return nil, errors.New("all routes exhausted")
}

func (g *Gateway) ExecuteStream(ctx context.Context, req *providers.Request) (<-chan providers.StreamEvent, error) {
	start := time.Now()
	prepared := g.PrepareRequest(req)

	targets, err := g.router.SelectTargets(prepared.Model)
	if err != nil {
		return nil, fmt.Errorf("route selection: %w", err)
	}

	var lastErr error

	for _, target := range targets {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		adapter, err := g.registry.Get(target.ProviderKind)
		if err != nil {
			lastErr = err
			continue
		}

		creds, err := g.credResolve(ctx, target.AccountID)
		if err != nil {
			lastErr = err
			continue
		}

		targetReq := *prepared
		targetReq.Model = target.ModelName

		streamChan, streamErr := adapter.ExecuteStream(ctx, &targetReq, creds)
		if streamErr == nil {
			g.cooldowns.MarkSuccess(target.AccountID)

			outChan := make(chan providers.StreamEvent, 16)
			go func(t routing.Target, in <-chan providers.StreamEvent) {
				defer close(outChan)
				var finalUsage providers.Usage

				for {
					select {
					case <-ctx.Done():
						select {
						case outChan <- providers.StreamEvent{Type: providers.StreamEventError, Error: ctx.Err()}:
						default:
						}
						g.recordUsage(req.ID, t, finalUsage, time.Since(start), 499, "client cancelled")
						return
					case ev, ok := <-in:
						if !ok {
							g.recordUsage(req.ID, t, finalUsage, time.Since(start), 200, "")
							return
						}
						if ev.Type == providers.StreamEventUsage && ev.Usage != nil {
							finalUsage = *ev.Usage
						}
						select {
						case outChan <- ev:
						case <-ctx.Done():
							g.recordUsage(req.ID, t, finalUsage, time.Since(start), 499, "client cancelled")
							return
						}
					}
				}
			}(target, streamChan)

			return outChan, nil
		}

		lastErr = streamErr

		if !providers.IsTransient(streamErr) {
			g.recordUsage(req.ID, target, providers.Usage{}, time.Since(start), 400, streamErr.Error())
			return nil, streamErr
		}

		g.cooldowns.MarkFailure(target.AccountID, 15*time.Second)
	}

	g.recordUsage(req.ID, routing.Target{}, providers.Usage{}, time.Since(start), 503, "all stream routes exhausted")
	if lastErr != nil {
		return nil, fmt.Errorf("all stream routes exhausted, last error: %w", lastErr)
	}
	return nil, errors.New("all stream routes exhausted")
}

func (g *Gateway) recordUsage(reqID string, t routing.Target, u providers.Usage, duration time.Duration, status int, errClass string) {
	if g.recorder == nil {
		return
	}
	g.recorder.Record(usage.UsageRecord{
		RequestID:    reqID,
		ProviderID:   t.ProviderID,
		AccountID:    t.AccountID,
		ModelID:      t.ModelName,
		InputTokens:  u.PromptTokens,
		OutputTokens: u.CompletionTokens,
		TotalTokens:  u.TotalTokens,
		LatencyMs:    int(duration.Milliseconds()),
		Status:       status,
		ErrorClass:   errClass,
	})
}
