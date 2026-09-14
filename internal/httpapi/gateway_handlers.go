package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dresar/ekarouter/internal/gateway"
	"github.com/dresar/ekarouter/internal/providers"
)

type GatewayHandler struct {
	gw *gateway.Gateway
	db *sql.DB
}

func NewGatewayHandler(gw *gateway.Gateway, db *sql.DB) *GatewayHandler {
	return &GatewayHandler{
		gw: gw,
		db: db,
	}
}

func (h *GatewayHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
SELECT external_name FROM models WHERE enabled = 1
UNION
SELECT name FROM routes WHERE enabled = 1`)
	if err != nil {
		http.Error(w, `{"error":"failed to query models"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ModelItem struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	var models []ModelItem
	now := time.Now().Unix()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			models = append(models, ModelItem{
				ID:      name,
				Object:  "model",
				Created: now,
				OwnedBy: "ekarouter",
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"object": "list",
		"data":   models,
	})
}

func (h *GatewayHandler) ChatCompletions(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"failed to read request body"}`, http.StatusBadRequest)
		return
	}

	var req providers.Request
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid json body: %v"}`, err), http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		var rawBody map[string]json.RawMessage
		_ = json.Unmarshal(bodyBytes, &rawBody)
		if inputRaw, ok := rawBody["input"]; ok {
			var inputStr string
			if err := json.Unmarshal(inputRaw, &inputStr); err == nil && inputStr != "" {
				req.Messages = []providers.Message{{Role: "user", Content: inputStr}}
			}
		}
	}

	if r.Header.Get("X-Token-Saver") == "off" {
		req.OptOutTokenSaver = true
	}

	if req.Model == "" {
		http.Error(w, `{"error":"model is required"}`, http.StatusBadRequest)
		return
	}

	reqID := GetRequestID(r.Context())
	if req.ID == "" {
		req.ID = reqID
	}

	if req.Stream {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, `{"error":"streaming unsupported"}`, http.StatusInternalServerError)
			return
		}

		streamChan, err := h.gw.ExecuteStream(r.Context(), &req)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		created := time.Now().Unix()

		for ev := range streamChan {
			switch ev.Type {
			case providers.StreamEventDelta:
				chunk := map[string]any{
					"id":      req.ID,
					"object":  "chat.completion.chunk",
					"created": created,
					"model":   req.Model,
					"choices": []map[string]any{
						{
							"index": 0,
							"delta": map[string]string{
								"content": ev.Delta,
							},
						},
					},
				}
				bytes, _ := json.Marshal(chunk)
				fmt.Fprintf(w, "data: %s\n\n", bytes)
				flusher.Flush()

			case providers.StreamEventReasoning:
				chunk := map[string]any{
					"id":      req.ID,
					"object":  "chat.completion.chunk",
					"created": created,
					"model":   req.Model,
					"choices": []map[string]any{
						{
							"index": 0,
							"delta": map[string]string{
								"reasoning_content": ev.Reasoning,
							},
						},
					},
				}
				bytes, _ := json.Marshal(chunk)
				fmt.Fprintf(w, "data: %s\n\n", bytes)
				flusher.Flush()

			case providers.StreamEventUsage:
				if ev.Usage != nil {
					chunk := map[string]any{
						"id":      req.ID,
						"object":  "chat.completion.chunk",
						"created": created,
						"model":   req.Model,
						"choices": []any{},
						"usage": map[string]int{
							"prompt_tokens":     ev.Usage.PromptTokens,
							"completion_tokens": ev.Usage.CompletionTokens,
							"total_tokens":      ev.Usage.TotalTokens,
						},
					}
					bytes, _ := json.Marshal(chunk)
					fmt.Fprintf(w, "data: %s\n\n", bytes)
					flusher.Flush()
				}

			case providers.StreamEventDone:
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				return

			case providers.StreamEventError:
				errPayload := map[string]any{
					"error": ev.Error.Error(),
				}
				bytes, _ := json.Marshal(errPayload)
				fmt.Fprintf(w, "data: %s\n\n", bytes)
				flusher.Flush()
				return
			}
		}

		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
		return
	}

	resp, err := h.gw.Execute(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadGateway)
		return
	}

	openAIResponse := map[string]any{
		"id":      resp.ID,
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   resp.Model,
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]string{
					"role":    resp.Role,
					"content": resp.Content,
				},
				"finish_reason": resp.FinishReason,
			},
		},
		"usage": map[string]int{
			"prompt_tokens":     resp.Usage.PromptTokens,
			"completion_tokens": resp.Usage.CompletionTokens,
			"total_tokens":      resp.Usage.TotalTokens,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(openAIResponse)
}

func (h *GatewayHandler) Responses(w http.ResponseWriter, r *http.Request) {
	h.ChatCompletions(w, r)
}
