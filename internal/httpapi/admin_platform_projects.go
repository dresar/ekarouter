package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dresar/ekarouter/internal/rbac"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PlatformHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	list, err := h.rbacService.ListProjects(r.Context())
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "database_error", err.Error(), nil)
		return
	}
	if list == nil {
		list = []*rbac.Project{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Environment string `json:"environment"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Project name is required", nil)
		return
	}

	p, err := h.rbacService.CreateProject(r.Context(), body.Name, "admin", body.Environment, body.Description)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "project_error", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, p)
}

func (h *PlatformHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.rbacService.GetProject(r.Context(), id)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found", nil)
		return
	}
	h.writeSuccess(w, r, p)
}

func (h *PlatformHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name        string `json:"name"`
		Environment string `json:"environment"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_json", "Malformed JSON", nil)
		return
	}
	if err := h.rbacService.UpdateProject(r.Context(), id, body.Name, body.Environment, body.Description); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "update_error", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{"updated": true, "id": id})
}

func (h *PlatformHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.rbacService.DeleteProject(r.Context(), id); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "delete_error", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{"deleted": true, "id": id})
}

func (h *PlatformHandler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), "SELECT id, name, COALESCE(project_id, ''), description, created_at FROM environments ORDER BY created_at DESC")
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var id, name, pID, desc string
		var ca time.Time
		_ = rows.Scan(&id, &name, &pID, &desc, &ca)
		list = append(list, map[string]any{
			"id":          id,
			"name":        name,
			"project_id":  pID,
			"description": desc,
			"created_at":  ca,
		})
	}
	if list == nil {
		list = []map[string]any{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		ProjectID   string `json:"project_id"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Environment name is required", nil)
		return
	}

	envID := uuid.New().String()
	var pID any
	if body.ProjectID != "" {
		pID = body.ProjectID
	}
	_, err := h.db.ExecContext(r.Context(), `
INSERT INTO environments (id, name, project_id, description, created_at)
VALUES (?, ?, ?, ?, ?)`, envID, body.Name, pID, body.Description, time.Now().UTC())
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}

	h.writeSuccess(w, r, map[string]any{
		"id":          envID,
		"name":        body.Name,
		"project_id":  body.ProjectID,
		"description": body.Description,
	})
}

func (h *PlatformHandler) UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_json", "Malformed JSON", nil)
		return
	}
	_, err := h.db.ExecContext(r.Context(), "UPDATE environments SET name = ?, description = ? WHERE id = ?", body.Name, body.Description, id)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "update_error", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{"updated": true, "id": id})
}

func (h *PlatformHandler) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.ExecContext(r.Context(), "DELETE FROM environments WHERE id = ?", id)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "delete_error", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{"deleted": true, "id": id})
}
