package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/providers"
)

func parseAnthropicSystem(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var sb strings.Builder
		for _, b := range blocks {
			if b.Text != "" {
				if sb.Len() > 0 {
					sb.WriteString("\n\n")
				}
				sb.WriteString(b.Text)
			}
		}
		return sb.String()
	}
	return ""
}

func parseAnthropicMessages(raw []json.RawMessage) []providers.Message {
	var msgs []providers.Message
	for _, rm := range raw {
		var item struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		}
		if err := json.Unmarshal(rm, &item); err != nil {
			continue
		}
		var textStr string
		if err := json.Unmarshal(item.Content, &textStr); err == nil {
			msgs = append(msgs, providers.Message{
				Role:    item.Role,
				Content: textStr,
			})
			continue
		}
		var blocks []struct {
			Type      string          `json:"type"`
			Text      string          `json:"text,omitempty"`
			ID        string          `json:"id,omitempty"`
			Name      string          `json:"name,omitempty"`
			Input     json.RawMessage `json:"input,omitempty"`
			ToolUseID string          `json:"tool_use_id,omitempty"`
			Content   any             `json:"content,omitempty"`
		}
		if err := json.Unmarshal(item.Content, &blocks); err == nil {
			var sb strings.Builder
			var toolCalls []providers.ToolCall
			for _, b := range blocks {
				switch b.Type {
				case "text":
					if sb.Len() > 0 {
						sb.WriteString("\n")
					}
					sb.WriteString(b.Text)
				case "tool_use":
					toolCalls = append(toolCalls, providers.ToolCall{
						ID:   b.ID,
						Type: "function",
						Function: providers.FunctionCall{
							Name:      b.Name,
							Arguments: string(b.Input),
						},
					})
				case "tool_result":
					if sb.Len() > 0 {
						sb.WriteString("\n")
					}
					switch cv := b.Content.(type) {
					case string:
						sb.WriteString(cv)
					default:
						resBytes, _ := json.Marshal(cv)
						sb.WriteString(string(resBytes))
					}
				}
			}
			msgs = append(msgs, providers.Message{
				Role:      item.Role,
				Content:   sb.String(),
				ToolCalls: toolCalls,
			})
		}
	}
	return msgs
}

func formatAnthropicContent(resp *providers.Response) []any {
	var content []any
	if resp.Content != "" {
		content = append(content, map[string]any{
			"type": "text",
			"text": resp.Content,
		})
	}
	for _, tc := range resp.ToolCalls {
		var inputMap any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &inputMap); err != nil {
			inputMap = map[string]any{}
		}
		content = append(content, map[string]any{
			"type":  "tool_use",
			"id":    tc.ID,
			"name":  tc.Function.Name,
			"input": inputMap,
		})
	}
	if len(content) == 0 {
		content = append(content, map[string]any{
			"type": "text",
			"text": "",
		})
	}
	return content
}

func (h *GatewayHandler) Messages(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model     string            `json:"model"`
		Messages  []json.RawMessage `json:"messages"`
		MaxTokens int               `json:"max_tokens"`
		System    json.RawMessage   `json:"system"`
		Stream    bool              `json:"stream"`
		Tools     []any             `json:"tools"`
		Thinking  *struct {
			Type         string `json:"type"`
			BudgetTokens int    `json:"budget_tokens"`
		} `json:"thinking"`
	}
	if err := readJSON(r, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"type": "error", "error": map[string]string{"type": "invalid_request_error", "message": "invalid json"}})
		return
	}
	if req.Model == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"type": "error", "error": map[string]string{"type": "invalid_request_error", "message": "model is required"}})
		return
	}
	if len(req.Messages) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"type": "error", "error": map[string]string{"type": "invalid_request_error", "message": "messages is required"}})
		return
	}

	msgs := parseAnthropicMessages(req.Messages)
	systemStr := parseAnthropicSystem(req.System)
	if systemStr != "" {
		msgs = append([]providers.Message{{Role: "system", Content: systemStr}}, msgs...)
	}

	openAIReq := &providers.Request{
		Model:    req.Model,
		Messages: msgs,
		Stream:   req.Stream,
		Tools:    req.Tools,
	}
	if req.MaxTokens > 0 {
		openAIReq.MaxTokens = &req.MaxTokens
	}
	if req.Thinking != nil && req.Thinking.BudgetTokens > 0 {
		openAIReq.ThinkingBudget = &req.Thinking.BudgetTokens
	}

	if req.Stream {
		h.streamAnthropicMessages(w, r, openAIReq)
		return
	}

	resp, err := h.gw.Execute(r.Context(), openAIReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{"type": "error", "error": map[string]string{"type": "server_error", "message": err.Error()}})
		return
	}

	stopReason := resp.FinishReason
	if len(resp.ToolCalls) > 0 {
		stopReason = "tool_use"
	} else if stopReason == "" || stopReason == "stop" {
		stopReason = "end_turn"
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":            resp.ID,
		"type":          "message",
		"role":          "assistant",
		"content":       formatAnthropicContent(resp),
		"model":         resp.Model,
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"usage": map[string]int{
			"input_tokens":  resp.Usage.PromptTokens,
			"output_tokens": resp.Usage.CompletionTokens,
		},
	})
}

func (h *GatewayHandler) streamAnthropicMessages(w http.ResponseWriter, r *http.Request, req *providers.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"type": "error", "error": map[string]string{"message": "streaming unsupported"}})
		return
	}
	streamChan, err := h.gw.ExecuteStream(r.Context(), req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{"type": "error", "error": map[string]string{"message": err.Error()}})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	msgID := req.ID
	if msgID == "" {
		msgID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}

	startData, _ := json.Marshal(map[string]any{
		"type": "message_start",
		"message": map[string]any{
			"id": msgID, "type": "message", "role": "assistant",
			"content": []any{}, "model": req.Model,
		},
	})
	fmt.Fprintf(w, "event: message_start\ndata: %s\n\n", startData)

	cbStart, _ := json.Marshal(map[string]any{
		"type": "content_block_start", "index": 0,
		"content_block": map[string]string{"type": "text", "text": ""},
	})
	fmt.Fprintf(w, "event: content_block_start\ndata: %s\n\n", cbStart)
	flusher.Flush()

	for ev := range streamChan {
		switch ev.Type {
		case providers.StreamEventDelta:
			deltaData, _ := json.Marshal(map[string]any{
				"type": "content_block_delta", "index": 0,
				"delta": map[string]string{"type": "text_delta", "text": ev.Delta},
			})
			fmt.Fprintf(w, "event: content_block_delta\ndata: %s\n\n", deltaData)
			flusher.Flush()
		case providers.StreamEventDone:
			cbStop, _ := json.Marshal(map[string]string{"type": "content_block_stop"})
			fmt.Fprintf(w, "event: content_block_stop\ndata: %s\n\n", cbStop)
			msgStop, _ := json.Marshal(map[string]string{"type": "message_stop"})
			fmt.Fprintf(w, "event: message_stop\ndata: %s\n\n", msgStop)
			flusher.Flush()
			return
		case providers.StreamEventError:
			errData, _ := json.Marshal(map[string]any{
				"type": "error",
				"error": map[string]string{"type": "server_error", "message": ev.Error.Error()},
			})
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", errData)
			flusher.Flush()
			return
		}
	}
}

func (h *GatewayHandler) Search(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model      string `json:"model"`
		Provider   string `json:"provider"`
		Query      string `json:"query"`
		MaxResults int    `json:"max_results"`
		SearchType string `json:"search_type"`
		Country    string `json:"country"`
		Language   string `json:"language"`
		TimeRange  string `json:"time_range"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "invalid json body", http.StatusBadRequest)
		return
	}
	providerModel := req.Model
	if providerModel == "" {
		providerModel = req.Provider
	}
	if providerModel == "" {
		jsonError(w, "model (or provider) is required", http.StatusBadRequest)
		return
	}
	if req.Query == "" {
		jsonError(w, "query is required", http.StatusBadRequest)
		return
	}
	if req.MaxResults == 0 {
		req.MaxResults = 10
	}

	fwd := NewProxyForwarder(h.gw, h.db)
	resp, err := fwd.ForwardJSON(r.Context(), http.MethodPost, "/v1/search", req, providerModel)
	if err != nil {
		jsonError(w, "no upstream available: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
