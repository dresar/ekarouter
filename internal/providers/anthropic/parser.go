package anthropic

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
)

type ContentBlock struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	Thinking string          `json:"thinking,omitempty"`
	ID       string          `json:"id,omitempty"`
	Name     string          `json:"name,omitempty"`
	Input    json.RawMessage `json:"input,omitempty"`
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
		snippet := strings.TrimSpace(string(data))
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		return nil, fmt.Errorf("upstream returned invalid JSON response: %q (error: %w)", snippet, err)
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

func ExtractBlocks(blocks []ContentBlock) (string, string, []providers.ToolCall) {
	var text, thinking string
	var toolCalls []providers.ToolCall
	for _, b := range blocks {
		switch b.Type {
		case "text":
			text += b.Text
		case "thinking":
			thinking += b.Thinking
		case "tool_use":
			toolName := DecloakToolName(b.Name)
			toolCalls = append(toolCalls, providers.ToolCall{
				ID:   b.ID,
				Type: "function",
				Function: providers.FunctionCall{
					Name:      toolName,
					Arguments: string(b.Input),
				},
			})
		}
	}
	return text, thinking, toolCalls
}

func ConvertUsage(inputTokens, outputTokens int) providers.Usage {
	return providers.Usage{
		PromptTokens:     inputTokens,
		CompletionTokens: outputTokens,
		TotalTokens:      inputTokens + outputTokens,
	}
}
