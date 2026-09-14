package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dresar/ekarouter/internal/devtools"
)

func (a *AdminHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	ts := devtools.NewTemplateStore(a.db)
	templates, err := ts.ListProviderTemplates(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list templates"}`, http.StatusInternalServerError)
		return
	}

	reqTemplates, err := ts.ListRequestTemplates(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list request templates"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"provider_templates": templates,
		"request_templates":  reqTemplates,
	})
}

func (a *AdminHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	ts := devtools.NewTemplateStore(a.db)

	var body devtools.RequestTemplate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	if err := ts.CreateRequestTemplate(r.Context(), &body); err != nil {
		http.Error(w, `{"error":"failed to create template"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(body)
}

func (a *AdminHandler) ImportOpenAPI(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL     string `json:"url"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	var docBytes []byte

	if body.Content != "" {
		docBytes = []byte(body.Content)
	} else if body.URL != "" {
		if err := devtools.ValidateRemoteURL(body.URL); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		resp, err := http.Get(body.URL)
		if err != nil {
			http.Error(w, `{"error":"failed to fetch openapi document"}`, http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		docBytes, err = io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
		if err != nil {
			http.Error(w, `{"error":"failed to read openapi document"}`, http.StatusBadGateway)
			return
		}
	} else {
		http.Error(w, `{"error":"url or content is required"}`, http.StatusBadRequest)
		return
	}

	result, err := devtools.ParseOpenAPIDocument(docBytes)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"preview":           true,
		"provider":          result.ProviderTemplate,
		"operations":        result.Operations,
		"auth_schemes":      result.AuthSchemes,
		"destructive_count": result.DestructiveCount,
		"total_count":       result.TotalCount,
		"message":           "review and confirm before saving",
	})
}

func (a *AdminHandler) ConfirmOpenAPIImport(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
		http.Error(w, `{"error":"content is required"}`, http.StatusBadRequest)
		return
	}

	result, err := devtools.ParseOpenAPIDocument([]byte(body.Content))
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if err := devtools.SaveOpenAPIImport(r.Context(), a.db, result); err != nil {
		http.Error(w, `{"error":"failed to save import"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":     "imported",
		"provider":   result.ProviderTemplate.ProviderName,
		"operations": result.TotalCount,
	})
}

func (a *AdminHandler) GenerateCurl(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Method        string `json:"method"`
		Path          string `json:"path"`
		BaseURL       string `json:"base_url"`
		Headers       string `json:"headers"`
		BodySchema    string `json:"body_schema"`
		CredentialRef string `json:"credential_ref"`
		TimeoutMs     int    `json:"timeout_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	rt := &devtools.RequestTemplate{
		Method:        body.Method,
		Path:          body.Path,
		Headers:       body.Headers,
		BodySchema:    body.BodySchema,
		CredentialRef: body.CredentialRef,
		TimeoutMs:     body.TimeoutMs,
	}

	curl := devtools.GenerateCurl(rt, body.BaseURL)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"curl": curl})
}
