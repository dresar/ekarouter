package openai

import (
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
)

func DefaultCapabilities() providers.Capabilities {
	return providers.Capabilities{
		Vision:        true,
		ToolCalling:   true,
		Reasoning:     false,
		Streaming:     true,
		SystemPrompt:  true,
		PromptCaching: true,
	}
}

func ModelCapabilities(modelID string) providers.Capabilities {
	caps := DefaultCapabilities()
	lower := strings.ToLower(modelID)
	if strings.HasPrefix(lower, "o1") || strings.HasPrefix(lower, "o3") || strings.Contains(lower, "deepseek-r1") || strings.Contains(lower, "deepseek-reasoner") {
		caps.Reasoning = true
	}
	if strings.HasPrefix(lower, "o1-preview") || strings.HasPrefix(lower, "o1-mini") {
		caps.ToolCalling = false
	}
	return caps
}

func DefaultModels() []providers.ModelInfo {
	return []providers.ModelInfo{
		{
			ID:           "gpt-4o",
			Name:         "GPT-4o",
			ContextLimit: 128000,
			Streaming:    true,
			Capabilities: ModelCapabilities("gpt-4o"),
		},
		{
			ID:           "gpt-4o-mini",
			Name:         "GPT-4o Mini",
			ContextLimit: 128000,
			Streaming:    true,
			Capabilities: ModelCapabilities("gpt-4o-mini"),
		},
		{
			ID:           "o1",
			Name:         "o1",
			ContextLimit: 200000,
			Streaming:    true,
			Capabilities: ModelCapabilities("o1"),
		},
		{
			ID:           "o3-mini",
			Name:         "o3-mini",
			ContextLimit: 200000,
			Streaming:    true,
			Capabilities: ModelCapabilities("o3-mini"),
		},
		{
			ID:           "gpt-4-turbo",
			Name:         "GPT-4 Turbo",
			ContextLimit: 128000,
			Streaming:    true,
			Capabilities: ModelCapabilities("gpt-4-turbo"),
		},
	}
}
