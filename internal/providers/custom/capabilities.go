package custom

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
		PromptCaching: false,
	}
}

func BackendCapabilities(backend string) providers.Capabilities {
	caps := DefaultCapabilities()
	switch strings.ToLower(backend) {
	case BackendDeepSeek:
		caps.Reasoning = true
		caps.PromptCaching = true
	case BackendGroq:
		caps.Streaming = true
		caps.ToolCalling = true
	case BackendOllama:
		caps.PromptCaching = false
	case BackendOpenRouter:
		caps.Reasoning = true
		caps.PromptCaching = true
	}
	return caps
}
