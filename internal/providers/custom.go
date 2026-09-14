package providers

import (
	"context"
	"net/http"
)

type CustomAdapter struct {
	openAI *OpenAIAdapter
}

func NewCustomAdapter(client *http.Client) *CustomAdapter {
	return &CustomAdapter{
		openAI: NewOpenAIAdapter(client),
	}
}

func (a *CustomAdapter) Kind() string {
	return "custom"
}

func (a *CustomAdapter) Models(ctx context.Context, creds *Credentials) ([]ModelInfo, error) {
	return a.openAI.Models(ctx, creds)
}

func (a *CustomAdapter) Execute(ctx context.Context, req *Request, creds *Credentials) (*Response, error) {
	return a.openAI.Execute(ctx, req, creds)
}

func (a *CustomAdapter) ExecuteStream(ctx context.Context, req *Request, creds *Credentials) (<-chan StreamEvent, error) {
	return a.openAI.ExecuteStream(ctx, req, creds)
}
