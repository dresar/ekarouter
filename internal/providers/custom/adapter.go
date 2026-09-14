package custom

import (
	"context"
	"net/http"

	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/providers/openai"
)

type Adapter struct {
	backend string
	openAI  *openai.Adapter
}

func NewAdapter(client *http.Client) *Adapter {
	return NewBackendAdapter(BackendGeneric, client)
}

func NewBackendAdapter(backend string, client *http.Client) *Adapter {
	if backend == "" {
		backend = BackendGeneric
	}
	return &Adapter{
		backend: backend,
		openAI:  openai.NewAdapter(client),
	}
}

func (a *Adapter) Kind() string {
	return a.backend
}

func (a *Adapter) ensureCreds(creds *providers.Credentials) *providers.Credentials {
	resolvedBase := ResolveBaseURL(a.backend, creds)
	if creds == nil {
		return &providers.Credentials{
			BaseURL: resolvedBase,
		}
	}
	if creds.BaseURL == "" {
		cp := *creds
		cp.BaseURL = resolvedBase
		return &cp
	}
	return creds
}

func (a *Adapter) Models(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error) {
	effectiveCreds := a.ensureCreds(creds)
	models, err := a.openAI.Models(ctx, effectiveCreds)
	if err != nil {
		return nil, err
	}
	backendCaps := BackendCapabilities(a.backend)
	for i := range models {
		models[i].Capabilities = backendCaps
	}
	return models, nil
}

func (a *Adapter) Execute(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
	effectiveCreds := a.ensureCreds(creds)
	return a.openAI.Execute(ctx, req, effectiveCreds)
}

func (a *Adapter) ExecuteStream(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
	effectiveCreds := a.ensureCreds(creds)
	return a.openAI.ExecuteStream(ctx, req, effectiveCreds)
}
