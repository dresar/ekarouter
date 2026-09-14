package gemini

import "github.com/dresar/ekarouter/internal/providers"

func DefaultCapabilities() providers.Capabilities {
	return providers.Capabilities{
		Vision:        true,
		ToolCalling:   true,
		Reasoning:     true,
		Streaming:     true,
		SystemPrompt:  true,
		PromptCaching: true,
	}
}

func ModelCapabilities(modelID string) providers.Capabilities {
	caps := DefaultCapabilities()
	switch modelID {
	case "gemini-2.0-flash-lite", "gemini-2.0-flash-lite-preview-02-05":
		caps.Reasoning = false
		caps.PromptCaching = false
	case "gemini-1.5-flash-8b":
		caps.Reasoning = false
		caps.PromptCaching = false
	}
	return caps
}

func DefaultModels() []providers.ModelInfo {
	return []providers.ModelInfo{
		{
			ID:           "gemini-2.5-pro",
			Name:         "Gemini 2.5 Pro",
			ContextLimit: 2097152,
			Streaming:    true,
			Capabilities: ModelCapabilities("gemini-2.5-pro"),
		},
		{
			ID:           "gemini-2.5-flash",
			Name:         "Gemini 2.5 Flash",
			ContextLimit: 1048576,
			Streaming:    true,
			Capabilities: ModelCapabilities("gemini-2.5-flash"),
		},
		{
			ID:           "gemini-2.0-flash",
			Name:         "Gemini 2.0 Flash",
			ContextLimit: 1048576,
			Streaming:    true,
			Capabilities: ModelCapabilities("gemini-2.0-flash"),
		},
		{
			ID:           "gemini-2.0-flash-lite",
			Name:         "Gemini 2.0 Flash Lite",
			ContextLimit: 1048576,
			Streaming:    true,
			Capabilities: ModelCapabilities("gemini-2.0-flash-lite"),
		},
		{
			ID:           "gemini-1.5-pro",
			Name:         "Gemini 1.5 Pro",
			ContextLimit: 2097152,
			Streaming:    true,
			Capabilities: ModelCapabilities("gemini-1.5-pro"),
		},
		{
			ID:           "gemini-1.5-flash",
			Name:         "Gemini 1.5 Flash",
			ContextLimit: 1048576,
			Streaming:    true,
			Capabilities: ModelCapabilities("gemini-1.5-flash"),
		},
	}
}
