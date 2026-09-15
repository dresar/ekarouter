package gemini

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
)

type FunctionCallPart struct {
	Name string `json:"name"`
	Args any    `json:"args"`
}

type ContentPart struct {
	Text             string            `json:"text,omitempty"`
	Thought          bool              `json:"thought,omitempty"`
	ThoughtSignature string            `json:"thoughtSignature,omitempty"`
	FunctionCall     *FunctionCallPart `json:"functionCall,omitempty"`
}

type CandidateContent struct {
	Role  string        `json:"role"`
	Parts []ContentPart `json:"parts"`
}

type Candidate struct {
	Content      CandidateContent `json:"content"`
	FinishReason string           `json:"finishReason"`
}

type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type ResponsePayload struct {
	Candidates    []Candidate    `json:"candidates"`
	UsageMetadata *UsageMetadata `json:"usageMetadata,omitempty"`
}

func NormalizeFinishReason(r string) string {
	switch strings.ToUpper(r) {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "SAFETY":
		return "content_filter"
	case "RECITATION":
		return "content_filter"
	default:
		return strings.ToLower(r)
	}
}

func ParseResponsePayload(data []byte) (*ResponsePayload, error) {
	var payload ResponsePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		snippet := strings.TrimSpace(string(data))
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		return nil, fmt.Errorf("upstream returned invalid JSON response: %q (error: %w)", snippet, err)
	}
	return &payload, nil
}

func ExtractCandidateTextAndReasoning(cand Candidate) (string, string) {
	var text strings.Builder
	var reasoning strings.Builder

	for _, p := range cand.Content.Parts {
		if p.Thought {
			reasoning.WriteString(p.Text)
		} else {
			text.WriteString(p.Text)
		}
	}
	return text.String(), reasoning.String()
}

func ExtractUsage(meta *UsageMetadata) *providers.Usage {
	if meta == nil {
		return nil
	}
	return &providers.Usage{
		PromptTokens:     meta.PromptTokenCount,
		CompletionTokens: meta.CandidatesTokenCount,
		TotalTokens:      meta.TotalTokenCount,
	}
}
