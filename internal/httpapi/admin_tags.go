package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Tag struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	CreatedAt   string `json:"created_at"`
}

func (h *AdminHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `SELECT id, name, COALESCE(description,''), COALESCE(color,''), created_at FROM tags ORDER BY name`)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var tags []Tag
	for rows.Next() {
		var t Tag
		_ = rows.Scan(&t.ID, &t.Name, &t.Description, &t.Color, &t.CreatedAt)
		tags = append(tags, t)
	}
	if tags == nil {
		tags = []Tag{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": tags})
}

func (h *AdminHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var t Tag
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if t.Name == "" {
		http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
		return
	}
	t.ID = uuid.NewString()
	if _, err := h.db.ExecContext(r.Context(), `INSERT INTO tags (id,name,description,color) VALUES (?,?,?,?)`, t.ID, t.Name, t.Description, t.Color); err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}

func (h *AdminHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := h.db.ExecContext(r.Context(), `DELETE FROM tags WHERE id=?`, id); err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
