package openrouter

import (
	"context"
	"net/http"

	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/providers/openai"
)

const (
	DefaultBaseURL   = "https://openrouter.ai/api/v1"
	RefererHeader    = "HTTP-Referer"
	RefererValue     = "https://github.com/dresar/ekarouter"
	TitleHeader      = "X-Title"
	TitleValue       = "EkaRouter"
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
	return "openrouter"
}

func (a *Adapter) Kind() string {
	return "openrouter"
}

func (a *Adapter) prepareCreds(creds *providers.Credentials) *providers.Credentials {
	res := &providers.Credentials{}
	if creds != nil {
		*res = *creds
	}
	if res.BaseURL == "" {
		res.BaseURL = DefaultBaseURL
	}
	headers := make(map[string]string)
	for k, v := range res.Headers {
		headers[k] = v
	}
	if headers[RefererHeader] == "" {
		headers[RefererHeader] = RefererValue
	}
	if headers[TitleHeader] == "" {
		headers[TitleHeader] = TitleValue
	}
	res.Headers = headers
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
