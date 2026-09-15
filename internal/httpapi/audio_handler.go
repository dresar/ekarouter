package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
)

func (h *GatewayHandler) AudioSpeech(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model          string  `json:"model"`
		Input          string  `json:"input"`
		Voice          string  `json:"voice"`
		ResponseFormat string  `json:"response_format"`
		Speed          float64 `json:"speed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if req.Model == "" {
		jsonError(w, "model is required", http.StatusBadRequest)
		return
	}
	if req.Input == "" {
		jsonError(w, "input is required", http.StatusBadRequest)
		return
	}
	if req.ResponseFormat == "" {
		req.ResponseFormat = "mp3"
	}
	if req.Voice == "" {
		req.Voice = "alloy"
	}

	fwd := NewProxyForwarder(h.gw, h.db)
	resp, err := fwd.ForwardJSON(r.Context(), http.MethodPost, "/v1/audio/speech", req, req.Model)
	if err != nil {
		jsonError(w, "no upstream available: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		switch req.ResponseFormat {
		case "opus":
			ct = "audio/opus"
		case "aac":
			ct = "audio/aac"
		case "flac":
			ct = "audio/flac"
		case "wav":
			ct = "audio/wav"
		case "pcm":
			ct = "audio/pcm"
		default:
			ct = "audio/mpeg"
		}
	}
	w.Header().Set("Content-Type", ct)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *GatewayHandler) AudioTranscriptions(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonError(w, "invalid multipart form", http.StatusBadRequest)
		return
	}
	model := r.FormValue("model")
	if model == "" {
		jsonError(w, "model is required", http.StatusBadRequest)
		return
	}
	if _, _, err := r.FormFile("file"); err != nil {
		jsonError(w, "file is required", http.StatusBadRequest)
		return
	}

	fwd := NewProxyForwarder(h.gw, h.db)
	resp, err := fwd.ForwardMultipart(r.Context(), r, "/v1/audio/transcriptions", model)
	if err != nil {
		jsonError(w, "no upstream available: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

var builtinVoices = []map[string]any{
	{"voice_id": "alloy", "name": "Alloy", "language": "en", "gender": "neutral"},
	{"voice_id": "echo", "name": "Echo", "language": "en", "gender": "male"},
	{"voice_id": "fable", "name": "Fable", "language": "en", "gender": "male"},
	{"voice_id": "onyx", "name": "Onyx", "language": "en", "gender": "male"},
	{"voice_id": "nova", "name": "Nova", "language": "en", "gender": "female"},
	{"voice_id": "shimmer", "name": "Shimmer", "language": "en", "gender": "female"},
	{"voice_id": "ash", "name": "Ash", "language": "en", "gender": "neutral"},
	{"voice_id": "ballad", "name": "Ballad", "language": "en", "gender": "neutral"},
	{"voice_id": "coral", "name": "Coral", "language": "en", "gender": "female"},
	{"voice_id": "sage", "name": "Sage", "language": "en", "gender": "neutral"},
	{"voice_id": "verse", "name": "Verse", "language": "en", "gender": "neutral"},
}

func (h *GatewayHandler) AudioVoices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"object": "list",
		"data":   builtinVoices,
	})
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"message": msg,
			"type":    "invalid_request_error",
		},
	})
}
