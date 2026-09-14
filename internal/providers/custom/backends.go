package custom

import (
	"net/http"
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
)

const (
	BackendGeneric    = "custom"
	BackendOllama     = "ollama"
	BackendVLLM       = "vllm"
	BackendDeepSeek   = "deepseek"
	BackendLocalAI    = "localai"
	BackendGroq       = "groq"
	BackendOpenRouter = "openrouter"
	BackendMistral    = "mistral"
	BackendTogether   = "together"
)

func DefaultBaseURL(backend string) string {
	switch strings.ToLower(backend) {
	case BackendOllama:
		return "http://localhost:11434/v1"
	case BackendVLLM:
		return "http://localhost:8000/v1"
	case BackendDeepSeek:
		return "https://api.deepseek.com/v1"
	case BackendLocalAI:
		return "http://localhost:8080/v1"
	case BackendGroq:
		return "https://api.groq.com/openai/v1"
	case BackendOpenRouter:
		return "https://openrouter.ai/api/v1"
	case BackendMistral:
		return "https://api.mistral.ai/v1"
	case BackendTogether:
		return "https://api.together.xyz/v1"
	default:
		return ""
	}
}

func ResolveBaseURL(backend string, creds *providers.Credentials) string {
	if creds != nil && creds.BaseURL != "" {
		return strings.TrimRight(creds.BaseURL, "/")
	}
	def := DefaultBaseURL(backend)
	if def != "" {
		return def
	}
	return "http://localhost:8080/v1"
}

func ApplyBackendHeaders(backend string, creds *providers.Credentials, h http.Header) {
	h.Set("Content-Type", "application/json")
	if strings.ToLower(backend) == BackendOpenRouter {
		h.Set("HTTP-Referer", "https://github.com/dresar/ekarouter")
		h.Set("X-Title", "EkaRouter")
	}

	if creds != nil {
		if creds.APIKey != "" {
			h.Set("Authorization", "Bearer "+creds.APIKey)
		} else if creds.AccessToken != "" {
			h.Set("Authorization", "Bearer "+creds.AccessToken)
		}
	}
}
