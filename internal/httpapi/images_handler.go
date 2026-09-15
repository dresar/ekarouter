package httpapi

import (
	"io"
	"net/http"
)

func (h *GatewayHandler) ImageGenerations(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model          string `json:"model"`
		Prompt         string `json:"prompt"`
		N              int    `json:"n"`
		Size           string `json:"size"`
		Quality        string `json:"quality"`
		Style          string `json:"style"`
		ResponseFormat string `json:"response_format"`
		User           string `json:"user"`
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
	if req.N == 0 {
		req.N = 1
	}
	if req.Size == "" {
		req.Size = "1024x1024"
	}
	if req.ResponseFormat == "" {
		req.ResponseFormat = "url"
	}

	fwd := NewProxyForwarder(h.gw, h.db)
	resp, err := fwd.ForwardJSON(r.Context(), http.MethodPost, "/v1/images/generations", req, req.Model)
	if err != nil {
		jsonError(w, "no upstream available: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
