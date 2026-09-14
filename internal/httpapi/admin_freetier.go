package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/dresar/ekarouter/internal/freetier"
	"github.com/go-chi/chi/v5"
)

func (a *AdminHandler) ListFreeTiers(w http.ResponseWriter, r *http.Request) {
	if a.catalogStore == nil {
		http.Error(w, `{"error":"catalog store not initialized"}`, http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	filter := &freetier.CatalogFilter{
		Category:   q.Get("category"),
		AuthMethod: q.Get("auth_method"),
		Status:     q.Get("status"),
	}

	entries, err := a.catalogStore.ListEntries(r.Context(), filter)
	if err != nil {
		http.Error(w, `{"error":"failed to list entries"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"entries": entries, "count": len(entries)})
}

func (a *AdminHandler) GetFreeTier(w http.ResponseWriter, r *http.Request) {
	if a.catalogStore == nil {
		http.Error(w, `{"error":"catalog store not initialized"}`, http.StatusInternalServerError)
		return
	}

	id := chi.URLParam(r, "id")
	entry, err := a.catalogStore.GetEntry(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"entry not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(entry)
}

func (a *AdminHandler) RefreshFreeTiers(w http.ResponseWriter, r *http.Request) {
	if a.catalogStore == nil {
		http.Error(w, `{"error":"catalog store not initialized"}`, http.StatusInternalServerError)
		return
	}

	entries, err := a.catalogStore.ListEntries(r.Context(), nil)
	if err != nil {
		http.Error(w, `{"error":"failed to list entries for refresh"}`, http.StatusInternalServerError)
		return
	}

	refreshed := 0
	for _, e := range entries {
		if e.Status == freetier.StatusVerified {
			continue
		}
		_ = a.catalogStore.UpdateStatus(r.Context(), e.ID, freetier.StatusUnverified)
		refreshed++
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "refresh_completed",
		"refreshed": refreshed,
		"total":     len(entries),
	})
}

func (a *AdminHandler) ListFreeTierCategories(w http.ResponseWriter, r *http.Request) {
	if a.catalogStore == nil {
		http.Error(w, `{"error":"catalog store not initialized"}`, http.StatusInternalServerError)
		return
	}

	cats, err := a.catalogStore.GetCategories(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list categories"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"categories": cats})
}

func (a *AdminHandler) ListVerifiedFreeTiers(w http.ResponseWriter, r *http.Request) {
	if a.catalogStore == nil {
		http.Error(w, `{"error":"catalog store not initialized"}`, http.StatusInternalServerError)
		return
	}

	entries, err := a.catalogStore.GetVerified(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list verified entries"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"entries": entries, "count": len(entries)})
}

func (a *AdminHandler) GetFreeTierSources(w http.ResponseWriter, r *http.Request) {
	if a.catalogStore == nil {
		http.Error(w, `{"error":"catalog store not initialized"}`, http.StatusInternalServerError)
		return
	}

	id := chi.URLParam(r, "id")
	sources, err := a.catalogStore.GetSources(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"failed to get sources"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"sources": sources})
}
