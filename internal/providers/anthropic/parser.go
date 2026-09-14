package anthropic

import (
	"encoding/json"

	"github.com/dresar/ekarouter/internal/providers"
)

type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

type MessageResponse struct {
	ID         string         `json:"id"`
	Model      string         `json:"model"`
	Role       string         `json:"role"`
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type StreamEventPayload struct {
	Type  string `json:"type"`
	Delta struct {
		Type     string `json:"type"`
		Text     string `json:"text,omitempty"`
		Thinking string `json:"thinking,omitempty"`
	} `json:"delta"`
	Usage *struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

func ParseMessageResponse(data []byte) (*MessageResponse, error) {
	var resp MessageResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func ParseStreamEventPayload(data []byte) (*StreamEventPayload, error) {
	var ev StreamEventPayload
	if err := json.Unmarshal(data, &ev); err != nil {
		return nil, err
	}
	return &ev, nil
}

func ExtractBlocks(blocks []ContentBlock) (string, string) {
	var text, thinking string
	for _, b := range blocks {
		switch b.Type {
		case "text":
			text += b.Text
		case "thinking":
			thinking += b.Thinking
		}
	}
	return text, thinking
}

func ConvertUsage(inputTokens, outputTokens int) providers.Usage {
	return providers.Usage{
		PromptTokens:     inputTokens,
		CompletionTokens: outputTokens,
		TotalTokens:      inputTokens + outputTokens,
	}
}
