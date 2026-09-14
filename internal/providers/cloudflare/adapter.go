package cloudflare

import (
	"context"
	"net/http"
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/providers/openai"
)

const (
	DefaultBaseURL = "https://api.cloudflare.com/client/v4/accounts/{accountId}/ai/v1"
)

type Adapter struct {
	openAI *openai.Adapter
}

func NewAdapter(client *http.Client) *Adapter {
	return &Adapter{
		openAI: openai.NewAdapter(client),
	}
}

func (a *Adapter) Name() string {
	return "cloudflare-ai"
}

func (a *Adapter) Kind() string {
	return "cloudflare-ai"
}

func (a *Adapter) prepareCreds(creds *providers.Credentials) *providers.Credentials {
	res := &providers.Credentials{}
	if creds != nil {
		*res = *creds
	}
	base := res.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	accID := res.ProjectID
	if accID != "" && strings.Contains(base, "{accountId}") {
		base = strings.ReplaceAll(base, "{accountId}", accID)
	}
	res.BaseURL = base
	return res
}

func (a *Adapter) Models(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error) {
	effectiveCreds := a.prepareCreds(creds)
	return a.openAI.Models(ctx, effectiveCreds)
}

func (a *Adapter) Execute(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
	effectiveCreds := a.prepareCreds(creds)
	return a.openAI.Execute(ctx, req, effectiveCreds)
}

func (a *Adapter) ExecuteStream(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
	effectiveCreds := a.prepareCreds(creds)
	return a.openAI.ExecuteStream(ctx, req, effectiveCreds)
}
