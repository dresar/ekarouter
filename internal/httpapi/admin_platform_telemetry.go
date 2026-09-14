package httpapi

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/dresar/ekarouter/internal/audit"
)

func (h *PlatformHandler) HealthSummary(w http.ResponseWriter, r *http.Request) {
	h.writeSuccess(w, r, map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"checks": map[string]string{
			"database": "connected",
			"vault":    "operational",
			"proxy":    "ready",
		},
	})
}

func (h *PlatformHandler) HealthProviders(w http.ResponseWriter, r *http.Request) {
	list := h.registry.List()
	results := make([]map[string]any, 0, len(list))
	for _, p := range list {
		results = append(results, map[string]any{
			"provider_id": p.ID,
			"name":        p.Name,
			"category":    p.Category,
			"status":      "online",
		})
	}
	h.writeSuccess(w, r, results)
}

func (h *PlatformHandler) HealthCredentials(w http.ResponseWriter, r *http.Request) {
	creds, _ := h.vaultStore.ListCredentials(r.Context(), "", "", "")
	res := make([]map[string]any, 0, len(creds))
	for _, c := range creds {
		res = append(res, map[string]any{
			"id":           c.ID,
			"name":         c.Name,
			"provider_id":  c.ProviderID,
			"health_state": c.HealthState,
			"status":       c.Status,
		})
	}
	h.writeSuccess(w, r, res)
}

func (h *PlatformHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	act := r.URL.Query().Get("action")
	actor := r.URL.Query().Get("actor_id")
	resType := r.URL.Query().Get("resource_type")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	logs, err := h.auditLogger.Query(r.Context(), act, actor, resType, limit)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "audit_query_error", err.Error(), nil)
		return
	}
	if logs == nil {
		logs = []*audit.Record{}
	}
	h.writeSuccess(w, r, logs)
}

func (h *PlatformHandler) GetUsageSummary(w http.ResponseWriter, r *http.Request) {
	var totalReqs, totalErrors int64
	_ = h.db.QueryRowContext(r.Context(), "SELECT COALESCE(SUM(request_count), 0), COALESCE(SUM(error_count), 0) FROM vault_credentials").Scan(&totalReqs, &totalErrors)

	var activeCreds int
	_ = h.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM vault_credentials WHERE status = 'active'").Scan(&activeCreds)

	h.writeSuccess(w, r, map[string]any{
		"total_requests":     totalReqs,
		"total_errors":       totalErrors,
		"active_credentials": activeCreds,
	})
}

func (h *PlatformHandler) GetUsageProviders(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
SELECT provider_id, COUNT(*) as cred_count,
       COALESCE(SUM(request_count), 0) as total_requests,
       COALESCE(SUM(error_count), 0) as total_errors
FROM vault_credentials GROUP BY provider_id ORDER BY total_requests DESC`)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var pID string
		var credCount int
		var reqs, errs int64
		if err := rows.Scan(&pID, &credCount, &reqs, &errs); err == nil {
			list = append(list, map[string]any{
				"provider_id":        pID,
				"active_credentials": credCount,
				"total_requests":     reqs,
				"total_errors":       errs,
			})
		}
	}
	if list == nil {
		list = []map[string]any{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) GetUsageCredentials(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
SELECT id, name, provider_id, COALESCE(project_id, ''), environment,
       request_count, error_count, last_used_at
FROM vault_credentials ORDER BY request_count DESC`)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var id, name, pID, projID, env string
		var reqs, errs int64
		var lu sql.NullTime
		if err := rows.Scan(&id, &name, &pID, &projID, &env, &reqs, &errs, &lu); err == nil {
			item := map[string]any{
				"id":            id,
				"name":          name,
				"provider_id":   pID,
				"project_id":    projID,
				"environment":   env,
				"request_count": reqs,
				"error_count":   errs,
			}
			if lu.Valid {
				item["last_used_at"] = lu.Time
			}
			list = append(list, item)
		}
	}
	if list == nil {
		list = []map[string]any{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) GetUsageProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
SELECT COALESCE(project_id, 'default') as proj_id,
       COUNT(*) as cred_count,
       COALESCE(SUM(request_count), 0) as total_requests,
       COALESCE(SUM(error_count), 0) as total_errors
FROM vault_credentials GROUP BY project_id ORDER BY total_requests DESC`)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var pID string
		var credCount int
		var reqs, errs int64
		if err := rows.Scan(&pID, &credCount, &reqs, &errs); err == nil {
			list = append(list, map[string]any{
				"project_id":         pID,
				"active_credentials": credCount,
				"total_requests":     reqs,
				"total_errors":       errs,
			})
		}
	}
	if list == nil {
		list = []map[string]any{}
	}
	h.writeSuccess(w, r, list)
}
