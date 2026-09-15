package httpapi

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *GatewayHandler) VideoGenerations(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model       string `json:"model"`
		Prompt      string `json:"prompt"`
		Duration    int    `json:"duration"`
		AspectRatio string `json:"aspect_ratio"`
		Resolution  string `json:"resolution"`
		Quality     string `json:"quality"`
		N           int    `json:"n"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if req.Model == "" {
		jsonError(w, "model is required", http.StatusBadRequest)
		return
	}
	if req.Prompt == "" {
		jsonError(w, "prompt is required", http.StatusBadRequest)
		return
	}

	fwd := NewProxyForwarder(h.gw, h.db)
	resp, err := fwd.ForwardJSON(r.Context(), http.MethodPost, "/v1/videos/generations", req, req.Model)
	if err != nil {
		jsonError(w, "no upstream available: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *GatewayHandler) VideoGet(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "xai"
	}

	fwd := NewProxyForwarder(h.gw, h.db)
	resp, err := fwd.ForwardJSON(r.Context(), http.MethodGet, "/v1/videos/"+videoID, nil, provider)
	if err != nil {
		jsonError(w, "no upstream available: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
