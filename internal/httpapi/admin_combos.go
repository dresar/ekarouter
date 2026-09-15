package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/dresar/ekarouter/internal/combos"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *AdminHandler) ListCombos(w http.ResponseWriter, r *http.Request) {
	svc := combos.NewService(h.db)
	items, err := svc.List()
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": items})
}

func (h *AdminHandler) GetCombo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	svc := combos.NewService(h.db)
	c, err := svc.Get(id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(c)
}

func (h *AdminHandler) CreateCombo(w http.ResponseWriter, r *http.Request) {
	var c combos.Combo
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if c.Name == "" {
		http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
		return
	}
	if c.Strategy == "" {
		c.Strategy = "fallback"
	}
	c.ID = uuid.NewString()
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO combos (id,name,description,strategy,sticky_limit,enabled) VALUES (?,?,?,?,?,1)`,
		c.ID, c.Name, c.Description, c.Strategy, c.StickyLimit); err != nil {
		http.Error(w, `{"error":"db error: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	for i, m := range c.Models {
		m.ID = uuid.NewString()
		m.ComboID = c.ID
		if m.Priority == 0 {
			m.Priority = len(c.Models) - i
		}
		if m.Weight == 0 {
			m.Weight = 1
		}
		if _, err := tx.Exec(`INSERT INTO combo_models (id,combo_id,model,priority,weight,enabled) VALUES (?,?,?,?,?,1)`,
			m.ID, c.ID, m.Model, m.Priority, m.Weight); err != nil {
			http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
}

func (h *AdminHandler) DeleteCombo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	svc := combos.NewService(h.db)
	if err := svc.Delete(id); err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
