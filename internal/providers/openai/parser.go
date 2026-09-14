package openai

import (
	"encoding/json"
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
)

type ChoiceMessage struct {
	Role             string `json:"role"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

type Choice struct {
	Index        int           `json:"index"`
	Message      ChoiceMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type StreamDelta struct {
	Role             string `json:"role,omitempty"`
	Content          string `json:"content,omitempty"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        StreamDelta `json:"delta"`
	FinishReason string      `json:"finish_reason"`
}

type UsagePayload struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletionResponse struct {
	ID      string        `json:"id"`
	Model   string        `json:"model"`
	Choices []Choice      `json:"choices"`
	Usage   *UsagePayload `json:"usage,omitempty"`
}

type ChatCompletionChunk struct {
	ID      string         `json:"id"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
	Usage   *UsagePayload  `json:"usage,omitempty"`
}

func ParseChatResponse(data []byte) (*ChatCompletionResponse, error) {
	var resp ChatCompletionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func ParseChatChunk(data []byte) (*ChatCompletionChunk, error) {
	var chunk ChatCompletionChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, err
	}
	return &chunk, nil
}

func ConvertUsage(u *UsagePayload) *providers.Usage {
	if u == nil {
		return nil
	}
	return &providers.Usage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	}
}

type ResponsesOutputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ResponsesOutputItem struct {
	Type    string                   `json:"type"`
	Role    string                   `json:"role"`
	Content []ResponsesOutputContent `json:"content"`
}

type ResponsesUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type ResponsesApiResponse struct {
	ID     string                `json:"id"`
	Model  string                `json:"model"`
	Status string                `json:"status"`
	Output []ResponsesOutputItem `json:"output"`
	Usage  *ResponsesUsage       `json:"usage"`
}

type ResponsesApiDelta struct {
	Type  string `json:"type"`
	Delta string `json:"delta"`
}

func ParseResponsesResponse(data []byte) (*ResponsesApiResponse, error) {
	var resp ResponsesApiResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func ParseResponsesChunk(data []byte) (*ResponsesApiDelta, error) {
	var chunk ResponsesApiDelta
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, err
	}
	return &chunk, nil
}

func ExtractResponsesOutputText(output []ResponsesOutputItem) string {
	var b strings.Builder
	for _, item := range output {
		for _, c := range item.Content {
			if c.Text != "" {
				b.WriteString(c.Text)
			}
		}
	}
	return b.String()
}

func ExtractResponsesUsage(u *ResponsesUsage) *providers.Usage {
	if u == nil {
		return nil
	}
	total := u.TotalTokens
	if total == 0 {
		total = u.InputTokens + u.OutputTokens
	}
	return &providers.Usage{
		PromptTokens:     u.InputTokens,
		CompletionTokens: u.OutputTokens,
		TotalTokens:      total,
	}
}
