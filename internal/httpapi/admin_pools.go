package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/dresar/ekarouter/internal/credpool"
	"github.com/go-chi/chi/v5"
)

func (a *AdminHandler) ListPools(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	ownerFilter := r.URL.Query().Get("owner_id")
	pools, err := a.poolStore.ListPools(r.Context(), ownerFilter)
	if err != nil {
		http.Error(w, `{"error":"failed to list pools"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"pools": pools})
}

func (a *AdminHandler) CreatePool(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	var body struct {
		Name        string `json:"name"`
		OwnerID     string `json:"owner_id"`
		ProviderID  string `json:"provider_id"`
		Environment string `json:"environment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	if body.OwnerID == "" {
		body.OwnerID = "admin"
	}
	if body.Environment == "" {
		body.Environment = "production"
	}

	pool := &credpool.Pool{
		Name:        body.Name,
		OwnerID:     body.OwnerID,
		ProviderID:  body.ProviderID,
		Environment: body.Environment,
	}

	if err := a.poolStore.CreatePool(r.Context(), pool); err != nil {
		http.Error(w, `{"error":"failed to create pool"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(pool)
}

func (a *AdminHandler) GetPool(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	id := chi.URLParam(r, "id")
	pool, err := a.poolStore.GetPool(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"pool not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pool)
}

func (a *AdminHandler) DeletePool(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	id := chi.URLParam(r, "id")
	if err := a.poolStore.DeletePool(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete pool"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (a *AdminHandler) ListPoolMembers(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	poolID := chi.URLParam(r, "id")
	members, err := a.poolStore.ListMembers(r.Context(), poolID)
	if err != nil {
		http.Error(w, `{"error":"failed to list members"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"members": members})
}

func (a *AdminHandler) AddPoolMember(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	poolID := chi.URLParam(r, "id")
	var body struct {
		AccountID string `json:"account_id"`
		Priority  int    `json:"priority"`
		Weight    int    `json:"weight"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AccountID == "" {
		http.Error(w, `{"error":"account_id is required"}`, http.StatusBadRequest)
		return
	}

	if body.Priority <= 0 {
		body.Priority = 10
	}
	if body.Weight <= 0 {
		body.Weight = 1
	}

	member := &credpool.Member{
		PoolID:    poolID,
		AccountID: body.AccountID,
		Priority:  body.Priority,
		Weight:    body.Weight,
	}

	if err := a.poolStore.AddMember(r.Context(), member); err != nil {
		http.Error(w, `{"error":"failed to add member"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(member)
}

func (a *AdminHandler) RemovePoolMember(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	mid := chi.URLParam(r, "mid")
	if err := a.poolStore.RemoveMember(r.Context(), mid); err != nil {
		http.Error(w, `{"error":"failed to remove member"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "removed"})
}

func (a *AdminHandler) GetPoolPolicy(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	poolID := chi.URLParam(r, "id")
	policy, err := a.poolStore.GetPolicy(r.Context(), poolID)
	if err != nil {
		http.Error(w, `{"error":"policy not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(policy)
}

func (a *AdminHandler) UpdatePoolPolicy(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	poolID := chi.URLParam(r, "id")
	var body credpool.RotationPolicy
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	body.PoolID = poolID
	if body.ID == "" {
		body.ID = "rp_" + poolID[:8]
	}

	if err := a.poolStore.SetPolicy(r.Context(), &body); err != nil {
		http.Error(w, `{"error":"failed to update policy"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (a *AdminHandler) PoolHealthCheck(w http.ResponseWriter, r *http.Request) {
	if a.healthChecker == nil {
		http.Error(w, `{"error":"health checker not initialized"}`, http.StatusInternalServerError)
		return
	}

	poolID := chi.URLParam(r, "id")
	results := a.healthChecker.CheckPool(r.Context(), poolID)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"results": results})
}

func (a *AdminHandler) PoolRotate(w http.ResponseWriter, r *http.Request) {
	if a.poolEngine == nil {
		http.Error(w, `{"error":"rotation engine not initialized"}`, http.StatusInternalServerError)
		return
	}

	poolID := chi.URLParam(r, "id")
	var body struct {
		RequestID string `json:"request_id"`
		MemberID  string `json:"member_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	member, err := a.poolEngine.Select(r.Context(), &credpool.SelectRequest{
		PoolID:    poolID,
		RequestID: body.RequestID,
		MemberID:  body.MemberID,
	})
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"selected": map[string]any{
			"member_id":  member.ID,
			"account_id": member.AccountID,
			"priority":   member.Priority,
			"weight":     member.Weight,
			"status":     member.Status,
		},
	})
}

func (a *AdminHandler) PausePool(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	id := chi.URLParam(r, "id")
	if err := a.poolStore.UpdatePoolStatus(r.Context(), id, credpool.PoolPaused); err != nil {
		http.Error(w, `{"error":"failed to pause pool"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "paused"})
}

func (a *AdminHandler) ResumePool(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	id := chi.URLParam(r, "id")
	if err := a.poolStore.UpdatePoolStatus(r.Context(), id, credpool.PoolActive); err != nil {
		http.Error(w, `{"error":"failed to resume pool"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "active"})
}

func (a *AdminHandler) PoolUsage(w http.ResponseWriter, r *http.Request) {
	if a.poolStore == nil {
		http.Error(w, `{"error":"pool store not initialized"}`, http.StatusInternalServerError)
		return
	}

	poolID := chi.URLParam(r, "id")
	members, err := a.poolStore.ListMembers(r.Context(), poolID)
	if err != nil {
		http.Error(w, `{"error":"failed to get usage"}`, http.StatusInternalServerError)
		return
	}

	type memberUsage struct {
		MemberID      string `json:"member_id"`
		AccountID     string `json:"account_id"`
		TotalRequests int    `json:"total_requests"`
		SuccessCount  int    `json:"success_count"`
		FailureCount  int    `json:"failure_count"`
		Status        string `json:"status"`
	}

	var usages []memberUsage
	totalReqs := 0
	totalSuccess := 0
	totalFail := 0
	for _, m := range members {
		usages = append(usages, memberUsage{
			MemberID:      m.ID,
			AccountID:     m.AccountID,
			TotalRequests: m.TotalRequests,
			SuccessCount:  m.SuccessCount,
			FailureCount:  m.FailureCount,
			Status:        string(m.Status),
		})
		totalReqs += m.TotalRequests
		totalSuccess += m.SuccessCount
		totalFail += m.FailureCount
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"pool_id":        poolID,
		"total_requests": totalReqs,
		"total_success":  totalSuccess,
		"total_failures": totalFail,
		"members":        usages,
	})
}
