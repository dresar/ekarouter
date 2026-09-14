package anthropic

import (
	"net/http"
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
)

func BuildBetaHeaders(hasThinking bool) string {
	features := []string{
		"prompt-caching-2024-07-31",
		"pdfs-2024-09-25",
		"output-128k-2025-02-19",
	}
	if hasThinking {
		features = append(features, "thinking-2024-11-20")
	}
	return strings.Join(features, ",")
}

func ApplyAnthropicHeaders(h http.Header, creds *providers.Credentials, hasThinking bool) {
	h.Set("Content-Type", "application/json")
	h.Set("anthropic-version", "2023-06-01")
	h.Set("anthropic-beta", BuildBetaHeaders(hasThinking))

	if creds != nil {
		if creds.APIKey != "" {
			h.Set("x-api-key", creds.APIKey)
		} else if creds.AccessToken != "" {
			h.Set("Authorization", "Bearer "+creds.AccessToken)
		}
	}
}

func ConcatenateSystemPrompts(messages []providers.Message) string {
	var systemPrompts []string
	for _, m := range messages {
		if strings.ToLower(m.Role) == "system" && m.Content != "" {
			systemPrompts = append(systemPrompts, m.Content)
		}
	}
	return strings.Join(systemPrompts, "\n\n")
}

func CloakToolName(name string) string {
	if strings.HasSuffix(name, "_ide") {
		return name
	}
	return name + "_ide"
}

func DecloakToolName(name string) string {
	return strings.TrimSuffix(name, "_ide")
}

func ConfigureThinking(body map[string]any, req *providers.Request) bool {
	if req.ThinkingBudget != nil && *req.ThinkingBudget > 0 {
		budget := *req.ThinkingBudget
		body["thinking"] = map[string]any{
			"type":          "enabled",
			"budget_tokens": budget,
		}
		body["temperature"] = 1.0
		if maxTokens, ok := body["max_tokens"].(int); ok && maxTokens <= budget {
			body["max_tokens"] = budget + 4096
		}
		return true
	}
	return false
}
