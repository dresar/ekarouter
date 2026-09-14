package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type Checker struct {
	db        *sql.DB
	startTime time.Time
}

func NewChecker(db *sql.DB) *Checker {
	return &Checker{
		db:        db,
		startTime: time.Now(),
	}
}

func (c *Checker) HealthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"status": "ok",
		"uptime": time.Since(c.startTime).String(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (c *Checker) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := c.db.PingContext(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
}
