package gemini

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
)

type CLIConfig struct {
	DefaultProjectID string
	BaseURL          string
}

func DefaultCLIConfig() CLIConfig {
	return CLIConfig{
		DefaultProjectID: "default-gemini-project",
		BaseURL:          "https://cloudaicompanion.googleapis.com/v1",
	}
}

func ResolveProjectID(req *providers.Request, creds *providers.Credentials, fallback string) string {
	if req != nil && req.ProjectID != "" {
		return req.ProjectID
	}
	if creds != nil && creds.ProjectID != "" {
		return creds.ProjectID
	}
	return fallback
}

func BuildCLIHeaders(creds *providers.Credentials) http.Header {
	h := make(http.Header)
	h.Set("Content-Type", "application/json")
	h.Set("User-Agent", "gemini-cli/1.0.0")
	h.Set("X-Goog-Api-Client", "gl-go/ gccl/")
	if creds != nil {
		if creds.AccessToken != "" {
			h.Set("Authorization", "Bearer "+creds.AccessToken)
		} else if creds.APIKey != "" {
			h.Set("Authorization", "Bearer "+creds.APIKey)
		}
	}
	return h
}

func BuildCLIURL(baseURL, model string) string {
	cleaned := strings.TrimRight(baseURL, "/")
	if cleaned == "" {
		cleaned = "https://cloudaicompanion.googleapis.com/v1"
	}
	return fmt.Sprintf("%s/models/%s:generateContent", cleaned, model)
}

func WrapCLIRequest(projectID, model string, body map[string]any) map[string]any {
	return map[string]any{
		"project": projectID,
		"model":   model,
		"request": body,
	}
}
