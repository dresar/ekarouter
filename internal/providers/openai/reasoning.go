package openai

import (
	"regexp"
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
)

var unicodePropertyRegex = regexp.MustCompile(`\\p\{[^}]+\}`)

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
			item["tool_call_id"] = m.ToolCallID
		}
		if len(m.ToolCalls) > 0 {
			calls := make([]map[string]any, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				calls[j] = map[string]any{
					"id":   tc.ID,
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

func BuildCodexInput(messages []providers.Message) []map[string]any {
	input := make([]map[string]any, len(messages))
	for i, m := range messages {
		role := m.Role
		if strings.ToLower(role) == "system" {
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
		input[i] = item
	}
	return input
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

func NormalizeCodexTools(tools []any) []map[string]any {
	var normalized []map[string]any
	for _, t := range tools {
		tm, ok := t.(map[string]any)
		if !ok {
			continue
		}
		tType, _ := tm["type"].(string)
		if tType == "" || tType == "function" {
			fn, _ := tm["function"].(map[string]any)
			name := ""
			desc := ""
			var params any
			if fn != nil {
				name, _ = fn["name"].(string)
				desc, _ = fn["description"].(string)
				params = fn["parameters"]
			} else {
				name, _ = tm["name"].(string)
				desc, _ = tm["description"].(string)
				params = tm["parameters"]
			}
			if name == "" {
				continue
			}
			if len(name) > 128 {
				name = name[:128]
			}
			cleanedParams := sanitizeSchemaPatterns(params)
			toolItem := map[string]any{
				"type":       "function",
				"name":       name,
				"parameters": cleanedParams,
			}
			if desc != "" {
				toolItem["description"] = desc
			}
			normalized = append(normalized, toolItem)
		}
	}
	return normalized
}

func sanitizeSchemaPatterns(v any) any {
	switch val := v.(type) {
	case map[string]any:
		res := make(map[string]any)
		for k, item := range val {
			if k == "pattern" {
				if s, ok := item.(string); ok {
					item = unicodePropertyRegex.ReplaceAllString(s, "")
				}
			}
			res[k] = sanitizeSchemaPatterns(item)
		}
		return res
	case []any:
		res := make([]any, len(val))
		for i, item := range val {
			res[i] = sanitizeSchemaPatterns(item)
		}
		return res
	default:
		return v
	}
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
