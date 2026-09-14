package usage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

type UsageRecord struct {
	RequestID    string
	ProviderID   string
	AccountID    string
	ModelID      string
	RouteID      string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	LatencyMs    int
	Status       int
	ErrorClass   string
	InputBytes   int
	OutputBytes  int
}

type Summary struct {
	TotalRequests int `json:"total_requests"`
	TotalTokens   int `json:"total_tokens"`
	PromptTokens  int `json:"prompt_tokens"`
	OutputTokens  int `json:"output_tokens"`
	AvgLatencyMs  int `json:"avg_latency_ms"`
}

type Recorder struct {
	db       *sql.DB
	ch       chan UsageRecord
	wg       sync.WaitGroup
	ctx      context.Context
	cancelFn context.CancelFunc
}

func NewRecorder(db *sql.DB, bufferSize int) *Recorder {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Recorder{
		db:       db,
		ch:       make(chan UsageRecord, bufferSize),
		ctx:      ctx,
		cancelFn: cancel,
	}

	r.wg.Add(1)
	go r.worker()

	return r
}

func (r *Recorder) Record(rec UsageRecord) {
	select {
	case r.ch <- rec:
	default:
	}
}

func (r *Recorder) Close() {
	r.cancelFn()
	close(r.ch)
	r.wg.Wait()
}

func (r *Recorder) worker() {
	defer r.wg.Done()

	for rec := range r.ch {
		r.persist(rec)
	}
}

func (r *Recorder) persist(rec UsageRecord) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _ = r.db.ExecContext(ctx, `
INSERT INTO usage_logs (request_id, provider_id, account_id, model_id, route_id, input_tokens, output_tokens, total_tokens, latency_ms, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.RequestID, rec.ProviderID, rec.AccountID, rec.ModelID, rec.RouteID,
		rec.InputTokens, rec.OutputTokens, rec.TotalTokens, rec.LatencyMs, rec.Status,
	)

	_, _ = r.db.ExecContext(ctx, `
INSERT INTO request_logs (request_id, provider_id, account_id, model, status, error_class, latency_ms, input_bytes, output_bytes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.RequestID, rec.ProviderID, rec.AccountID, rec.ModelID, rec.Status,
		rec.ErrorClass, rec.LatencyMs, rec.InputBytes, rec.OutputBytes,
	)
}

func (r *Recorder) PruneLogs(ctx context.Context, retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays).Format("2006-01-02 15:04:05")

	if _, err := r.db.ExecContext(ctx, "DELETE FROM usage_logs WHERE created_at < ?", cutoff); err != nil {
		return fmt.Errorf("prune usage_logs: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, "DELETE FROM request_logs WHERE created_at < ?", cutoff); err != nil {
		return fmt.Errorf("prune request_logs: %w", err)
	}
	return nil
}

func (r *Recorder) GetSummary(ctx context.Context) (*Summary, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT COUNT(1), COALESCE(SUM(total_tokens), 0), COALESCE(SUM(input_tokens), 0), COALESCE(SUM(output_tokens), 0), COALESCE(AVG(latency_ms), 0)
FROM usage_logs`)

	var s Summary
	var avgLatency float64
	if err := row.Scan(&s.TotalRequests, &s.TotalTokens, &s.PromptTokens, &s.OutputTokens, &avgLatency); err != nil {
		return nil, err
	}
	s.AvgLatencyMs = int(avgLatency)
	return &s, nil
}
