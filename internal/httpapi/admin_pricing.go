package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/dresar/ekarouter/internal/pricing"
	"github.com/go-chi/chi/v5"
)

func (h *AdminHandler) ListPricing(w http.ResponseWriter, r *http.Request) {
	svc := pricing.NewService(h.db)
	provider := r.URL.Query().Get("provider")
	items, err := svc.List(provider)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": items})
}

func (h *AdminHandler) UpsertPricing(w http.ResponseWriter, r *http.Request) {
	var p pricing.ModelPricing
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if p.ModelID == "" || p.Provider == "" {
		http.Error(w, `{"error":"model_id and provider required"}`, http.StatusBadRequest)
		return
	}
	svc := pricing.NewService(h.db)
	if err := svc.Upsert(p); err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *AdminHandler) DeletePricing(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	svc := pricing.NewService(h.db)
	if err := svc.Delete(id); err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) EstimatePricing(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model        string `json:"model"`
		InputTokens  int    `json:"input_tokens"`
		OutputTokens int    `json:"output_tokens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	svc := pricing.NewService(h.db)
	est, err := svc.EstimateCost(req.Model, req.InputTokens, req.OutputTokens)
	if err != nil {
		http.Error(w, `{"error":"estimation error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(est)
}
