package openai

import (
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
)

func IsReasoningModel(model string) bool {
	lower := strings.ToLower(model)
	return strings.HasPrefix(lower, "o1") || strings.HasPrefix(lower, "o3") || strings.Contains(lower, "deepseek-r1") || strings.Contains(lower, "deepseek-reasoner")
}

func AdaptMessagesForReasoning(messages []providers.Message, isReasoning bool) []map[string]any {
	adapted := make([]map[string]any, len(messages))
	for i, m := range messages {
		role := m.Role
		if isReasoning && strings.ToLower(role) == "system" {
			role = "developer"
		}
		item := map[string]any{
			"role":    role,
			"content": m.Content,
		}
		if m.Name != "" {
			item["name"] = m.Name
		}
		if m.ToolCallID != "" {
			item["tool_call_id"] = StripServerID(m.ToolCallID)
		}
		if len(m.ToolCalls) > 0 {
			calls := make([]map[string]any, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				calls[j] = map[string]any{
					"id":   StripServerID(tc.ID),
					"type": tc.Type,
					"function": map[string]any{
						"name":      tc.Function.Name,
						"arguments": tc.Function.Arguments,
					},
				}
			}
			item["tool_calls"] = calls
		}
		adapted[i] = item
	}
	return adapted
}

func StripServerID(id string) string {
	if strings.HasPrefix(id, "rs_") || strings.HasPrefix(id, "fc_") || strings.HasPrefix(id, "resp_") || strings.HasPrefix(id, "msg_") {
		parts := strings.SplitN(id, "_", 2)
		if len(parts) == 2 && len(parts[1]) > 0 {
			return parts[1]
		}
	}
	return id
}

func ApplyReasoningParameters(body map[string]any, req *providers.Request, isReasoning bool) {
	if !isReasoning {
		return
	}

	effort := strings.ToLower(req.ReasoningEffort)
	switch effort {
	case "low", "medium", "high":
		body["reasoning_effort"] = effort
	}

	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		body["max_completion_tokens"] = *req.MaxTokens
		delete(body, "max_tokens")
	}
}
