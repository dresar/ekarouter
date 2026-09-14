package anthropic

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
	if strings.Contains(lower, "3-7") || strings.Contains(lower, "thinking") {
		caps.Reasoning = true
	}
	return caps
}

func DefaultModels() []providers.ModelInfo {
	return []providers.ModelInfo{
		{
			ID:           "claude-3-7-sonnet-20250219",
			Name:         "Claude 3.7 Sonnet",
			ContextLimit: 200000,
			Streaming:    true,
			Capabilities: ModelCapabilities("claude-3-7-sonnet-20250219"),
		},
		{
			ID:           "claude-3-5-sonnet-20241022",
			Name:         "Claude 3.5 Sonnet",
			ContextLimit: 200000,
			Streaming:    true,
			Capabilities: ModelCapabilities("claude-3-5-sonnet-20241022"),
		},
		{
			ID:           "claude-3-5-haiku-20241022",
			Name:         "Claude 3.5 Haiku",
			ContextLimit: 200000,
			Streaming:    true,
			Capabilities: ModelCapabilities("claude-3-5-haiku-20241022"),
		},
		{
			ID:           "claude-3-opus-20240229",
			Name:         "Claude 3 Opus",
			ContextLimit: 200000,
			Streaming:    true,
			Capabilities: ModelCapabilities("claude-3-opus-20240229"),
		},
	}
}
