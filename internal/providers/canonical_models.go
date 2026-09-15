package providers

import "strings"

var CanonicalModels = map[string][]ModelInfo{
	"antigravity": {
		{ID: "gemini-3.8-flash-high", Name: "Gemini 3.8 Flash (High)", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gemini-3.8-flash-medium", Name: "Gemini 3.8 Flash (Medium)", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gemini-3.8-flash-low", Name: "Gemini 3.8 Flash (Low)", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gemini-3.8-flash", Name: "Gemini 3.8 Flash", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gemini-3.7-flash-high", Name: "Gemini 3.7 Flash (High)", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gemini-3.7-flash-medium", Name: "Gemini 3.7 Flash (Medium)", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gemini-3.7-flash", Name: "Gemini 3.7 Flash", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gemini-3.6-flash-high", Name: "Gemini 3.6 Flash (High)", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gemini-3.5-flash-high", Name: "Gemini 3.5 Flash (High)", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gemini-pro-agent", Name: "Gemini 3.1 Pro (High)", ContextLimit: 2097152, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gemini-3.1-pro-low", Name: "Gemini 3.1 Pro (Low)", ContextLimit: 2097152, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6 (Thinking)", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "claude-opus-4-6-thinking", Name: "Claude Opus 4.6 (Thinking)", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gpt-oss-120b-medium", Name: "GPT-OSS 120B (Medium)", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: false, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gemini-3.1-flash-image", Name: "Gemini 3.1 Flash (Image)", ContextLimit: 32000, Streaming: false, Capabilities: Capabilities{Vision: true, Streaming: false}},
		{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", ContextLimit: 2097152, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
	},
	"gemini": {
		{ID: "gemini-3.8-flash", Name: "Gemini 3.8 Flash", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gemini-3.7-flash", Name: "Gemini 3.7 Flash", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gemini-3.6-flash", Name: "Gemini 3.6 Flash", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gemini-3.5-flash-lite", Name: "Gemini 3.5 Flash Lite", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gemini-3.1-pro-preview", Name: "Gemini 3.1 Pro Preview", ContextLimit: 2097152, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", ContextLimit: 2097152, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "text-embedding-004", Name: "Text Embedding 004", ContextLimit: 2048, Streaming: false},
	},
	"anthropic": {
		{ID: "claude-3-7-sonnet-20250219", Name: "Claude 3.7 Sonnet", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "claude-opus-4-20250514", Name: "Claude Opus 4", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
	},
	"openai": {
		{ID: "gpt-4o", Name: "GPT-4o", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "o1", Name: "O1", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "o1-mini", Name: "O1 Mini", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: false, ToolCalling: false, Reasoning: true, Streaming: true}},
		{ID: "o3-mini", Name: "O3 Mini", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: false, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "gpt-4-turbo", Name: "GPT-4 Turbo", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "text-embedding-3-small", Name: "Text Embedding 3 Small", ContextLimit: 8191, Streaming: false},
		{ID: "text-embedding-3-large", Name: "Text Embedding 3 Large", ContextLimit: 8191, Streaming: false},
		{ID: "text-embedding-ada-002", Name: "Text Embedding Ada 002", ContextLimit: 8191, Streaming: false},
	},
	"groq": {
		{ID: "llama-3.3-70b-versatile", Name: "Llama 3.3 70B Versatile", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "llama-3.1-8b-instant", Name: "Llama 3.1 8B Instant", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "mixtral-8x7b-32768", Name: "Mixtral 8x7B", ContextLimit: 32768, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "deepseek-r1-distill-llama-70b", Name: "DeepSeek R1 Distill Llama 70B", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
	},
	"openrouter": {
		{ID: "deepseek/deepseek-r1", Name: "DeepSeek R1", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "deepseek/deepseek-chat", Name: "DeepSeek V3", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "meta-llama/llama-3.3-70b-instruct", Name: "Llama 3.3 70B Instruct", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "qwen/qwen-2.5-72b-instruct", Name: "Qwen 2.5 72B Instruct", ContextLimit: 32768, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
	},
	"kiro": {
		{ID: "claude-sonnet-4.5", Name: "Claude Sonnet 4.5", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-haiku-4.5", Name: "Claude Haiku 4.5", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
	},
	"opencode": {
		{ID: "gpt-4o", Name: "GPT-4o (OpenCode)", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-3-5-sonnet", Name: "Claude 3.5 Sonnet (OpenCode)", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
	},
	"ollama": {
		{ID: "llama3.2", Name: "Llama 3.2", ContextLimit: 8192, Streaming: true, Capabilities: Capabilities{Streaming: true}},
		{ID: "nomic-embed-text", Name: "Nomic Embed Text", ContextLimit: 2048, Streaming: false},
	},
}

func GetCanonicalModels(providerID string) []ModelInfo {
	lower := strings.ToLower(strings.TrimSpace(providerID))
	if models, ok := CanonicalModels[lower]; ok {
		return models
	}
	return nil
}

func GetAllCanonicalModels() []ModelInfo {
	var all []ModelInfo
	for _, list := range CanonicalModels {
		all = append(all, list...)
	}
	return all
}

func FindCanonicalModel(modelID string) *ModelInfo {
	target := strings.ToLower(strings.TrimSpace(modelID))
	for _, list := range CanonicalModels {
		for _, m := range list {
			if strings.ToLower(m.ID) == target {
				return &m
			}
		}
	}
	return nil
}
