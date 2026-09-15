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
		{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
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
		{ID: "claude-opus-4-20250514", Name: "Claude Opus 4", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "claude-sonnet-4-5-20250903", Name: "Claude Sonnet 4.5", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "claude-3-7-sonnet-20250219", Name: "Claude 3.7 Sonnet", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
	},
	"openai": {
		{ID: "gpt-4.1", Name: "GPT-4.1", ContextLimit: 1047576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gpt-4.1-mini", Name: "GPT-4.1 Mini", ContextLimit: 1047576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gpt-4.1-nano", Name: "GPT-4.1 Nano", ContextLimit: 1047576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gpt-4.5-preview", Name: "GPT-4.5 Preview", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gpt-4o", Name: "GPT-4o", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "o3", Name: "O3", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "o3-mini", Name: "O3 Mini", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: false, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "o4-mini", Name: "O4 Mini", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "o1", Name: "O1", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "text-embedding-3-small", Name: "Text Embedding 3 Small", ContextLimit: 8191, Streaming: false},
		{ID: "text-embedding-3-large", Name: "Text Embedding 3 Large", ContextLimit: 8191, Streaming: false},
		{ID: "text-embedding-ada-002", Name: "Text Embedding Ada 002", ContextLimit: 8191, Streaming: false},
		{ID: "tts-1", Name: "TTS-1", ContextLimit: 4096, Streaming: false},
		{ID: "tts-1-hd", Name: "TTS-1 HD", ContextLimit: 4096, Streaming: false},
		{ID: "whisper-1", Name: "Whisper-1", ContextLimit: 0, Streaming: false},
		{ID: "dall-e-3", Name: "DALL-E 3", ContextLimit: 4000, Streaming: false},
		{ID: "dall-e-2", Name: "DALL-E 2", ContextLimit: 1000, Streaming: false},
	},
	"groq": {
		{ID: "llama-4-scout-17b-16e-instruct", Name: "Llama 4 Scout 17B", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "llama-4-maverick-17b-128e-instruct", Name: "Llama 4 Maverick 17B", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "llama-3.3-70b-versatile", Name: "Llama 3.3 70B Versatile", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "llama-3.1-8b-instant", Name: "Llama 3.1 8B Instant", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "deepseek-r1-distill-llama-70b", Name: "DeepSeek R1 Distill Llama 70B", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "qwen-qwq-32b", Name: "QwQ 32B", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "mistral-saba-24b", Name: "Mistral Saba 24B", ContextLimit: 32768, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
	},
	"openrouter": {
		{ID: "anthropic/claude-opus-4", Name: "Claude Opus 4", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "anthropic/claude-sonnet-4-5", Name: "Claude Sonnet 4.5", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Reasoning: true, Streaming: true}},
		{ID: "openai/gpt-4.1", Name: "GPT-4.1", ContextLimit: 1047576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "x-ai/grok-3", Name: "Grok 3", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "x-ai/grok-3-mini", Name: "Grok 3 Mini", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "deepseek/deepseek-r1", Name: "DeepSeek R1", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "deepseek/deepseek-chat", Name: "DeepSeek V3", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "google/gemini-2.5-pro", Name: "Gemini 2.5 Pro", ContextLimit: 2097152, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "meta-llama/llama-4-maverick", Name: "Llama 4 Maverick", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "meta-llama/llama-3.3-70b-instruct", Name: "Llama 3.3 70B Instruct", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "qwen/qwen3-235b-a22b", Name: "Qwen3 235B A22B", ContextLimit: 40960, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "mistral/mistral-large-2411", Name: "Mistral Large 2411", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
	},
	"cerebras": {
		{ID: "llama-4-scout-17b-16e-instruct", Name: "Llama 4 Scout 17B", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "llama3.3-70b", Name: "Llama 3.3 70B", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "llama3.1-70b", Name: "Llama 3.1 70B", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "llama3.1-8b", Name: "Llama 3.1 8B", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "qwen-3-32b", Name: "Qwen 3 32B", ContextLimit: 32768, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
	},
	"nvidia": {
		{ID: "meta/llama-4-maverick-17b-128e-instruct", Name: "Llama 4 Maverick", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "meta/llama-3.3-70b-instruct", Name: "Llama 3.3 70B", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "deepseek-ai/deepseek-r1", Name: "DeepSeek R1", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "qwen/qwen3-235b-a22b", Name: "Qwen3 235B A22B", ContextLimit: 40960, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "google/gemma-3-27b-it", Name: "Gemma 3 27B", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "mistralai/mistral-large-2-instruct", Name: "Mistral Large 2", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
	},
	"kiro": {
		{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-sonnet-4.5", Name: "Claude Sonnet 4.5", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-haiku-4.5", Name: "Claude Haiku 4.5", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
	},
	"opencode": {
		{ID: "gpt-4.1", Name: "GPT-4.1 (OpenCode)", ContextLimit: 1047576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "gpt-4o", Name: "GPT-4o (OpenCode)", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5 (OpenCode)", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-3-5-sonnet", Name: "Claude 3.5 Sonnet (OpenCode)", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
	},
	"ollama": {
		{ID: "llama3.3", Name: "Llama 3.3", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "llama3.2", Name: "Llama 3.2", ContextLimit: 8192, Streaming: true, Capabilities: Capabilities{Streaming: true}},
		{ID: "llama3.1", Name: "Llama 3.1", ContextLimit: 32768, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "qwen2.5", Name: "Qwen 2.5", ContextLimit: 32768, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "qwen2.5-coder", Name: "Qwen 2.5 Coder", ContextLimit: 32768, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "phi4", Name: "Phi-4", ContextLimit: 16384, Streaming: true, Capabilities: Capabilities{Streaming: true}},
		{ID: "mistral", Name: "Mistral 7B", ContextLimit: 32768, Streaming: true, Capabilities: Capabilities{Streaming: true}},
		{ID: "deepseek-r1", Name: "DeepSeek R1", ContextLimit: 32768, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "gemma3", Name: "Gemma 3", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{Vision: true, Streaming: true}},
		{ID: "nomic-embed-text", Name: "Nomic Embed Text", ContextLimit: 2048, Streaming: false},
		{ID: "mxbai-embed-large", Name: "MxBai Embed Large", ContextLimit: 512, Streaming: false},
	},
	"huggingface": {
		{ID: "meta-llama/Llama-3.3-70B-Instruct", Name: "Llama 3.3 70B Instruct", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "meta-llama/Llama-4-Scout-17B-16E-Instruct", Name: "Llama 4 Scout 17B", ContextLimit: 131072, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "deepseek-ai/DeepSeek-R1", Name: "DeepSeek R1", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "Qwen/Qwen3-235B-A22B", Name: "Qwen3 235B A22B", ContextLimit: 40960, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "mistralai/Mistral-7B-Instruct-v0.3", Name: "Mistral 7B Instruct v0.3", ContextLimit: 32768, Streaming: true, Capabilities: Capabilities{Streaming: true}},
	},
	"chutes": {
		{ID: "deepseek-ai/DeepSeek-V3-0324", Name: "DeepSeek V3", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "deepseek-ai/DeepSeek-R1", Name: "DeepSeek R1", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "Qwen/Qwen3-235B-A22B", Name: "Qwen3 235B A22B", ContextLimit: 40960, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "meta-llama/Llama-4-Maverick-17B-128E-Instruct-FP8", Name: "Llama 4 Maverick", ContextLimit: 1048576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
	},
	"airforce": {
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini (Free)", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku (Free)", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "llama-3.3-70b-versatile", Name: "Llama 3.3 70B (Free)", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "deepseek-v3", Name: "DeepSeek V3 (Free)", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
	},
	"bazaarlink": {
		{ID: "gpt-4o", Name: "GPT-4o (BazaarLink)", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-3-5-sonnet", Name: "Claude 3.5 Sonnet (BazaarLink)", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
	},
	"kimchi": {
		{ID: "deepseek-r1", Name: "DeepSeek R1 (Kimchi)", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "deepseek-v3", Name: "DeepSeek V3 (Kimchi)", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini (Kimchi)", ContextLimit: 128000, Streaming: true, Capabilities: Capabilities{Vision: true, Streaming: true}},
	},
	"kilogateway": {
		{ID: "gpt-4.1-mini", Name: "GPT-4.1 Mini (Kilo)", ContextLimit: 1047576, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "claude-3-5-sonnet", Name: "Claude 3.5 Sonnet (Kilo)", ContextLimit: 200000, Streaming: true, Capabilities: Capabilities{Vision: true, ToolCalling: true, Streaming: true}},
		{ID: "deepseek-v3", Name: "DeepSeek V3 (Kilo)", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
	},
	"mimofree": {
		{ID: "qwen3-235b", Name: "Qwen3 235B (MiMo Free)", ContextLimit: 32768, Streaming: true, Capabilities: Capabilities{Reasoning: true, Streaming: true}},
		{ID: "deepseek-v3", Name: "DeepSeek V3 (MiMo Free)", ContextLimit: 64000, Streaming: true, Capabilities: Capabilities{ToolCalling: true, Streaming: true}},
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
