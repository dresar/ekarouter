package scheduler

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/dresar/ekarouter/internal/platform"
	"github.com/google/uuid"
)

type JobFunc func(ctx context.Context) error

type Scheduler struct {
	db            *sql.DB
	registry      *platform.Registry
	retentionDays int
	interval      time.Duration
	stopCh        chan struct{}
	wg            sync.WaitGroup
	mu            sync.Mutex
	running       bool
}

func NewScheduler(db *sql.DB, registry *platform.Registry, retentionDays int, interval time.Duration) *Scheduler {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &Scheduler{
		db:            db,
		registry:      registry,
		retentionDays: retentionDays,
		interval:      interval,
		stopCh:        make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go s.loop(ctx)
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()
	s.wg.Wait()
}

func (s *Scheduler) loop(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.runOnce(ctx)

	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOnce(ctx)
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context) {
	_ = s.ExecuteTask(ctx, "cleanup_cooldowns", s.cleanupCooldowns)
	_ = s.ExecuteTask(ctx, "prune_logs", s.pruneLogs)
}

func (s *Scheduler) ExecuteTask(ctx context.Context, taskType string, fn JobFunc) error {
	runID := uuid.New().String()
	now := time.Now().UTC()
	taskID := taskType

	_, err := s.db.ExecContext(ctx, `
INSERT INTO scheduled_tasks (id, task_type, target_id, cron_expr, status, last_run_at, created_at)
VALUES (?, ?, '', '@interval', 'active', ?, ?)
ON CONFLICT(id) DO UPDATE SET last_run_at = excluded.last_run_at`, taskID, taskType, now, now)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
INSERT INTO task_runs (id, task_id, status, details, started_at)
VALUES (?, ?, 'running', '', ?)`, runID, taskID, now)
	if err != nil {
		return err
	}

	jobErr := fn(ctx)
	finished := time.Now().UTC()
	status := "completed"
	details := ""
	if jobErr != nil {
		status = "failed"
		details = jobErr.Error()
	}

	_, _ = s.db.ExecContext(ctx, `
UPDATE task_runs
SET status = ?, details = ?, finished_at = ?
WHERE id = ?`, status, details, finished, runID)

	return jobErr
}

func (s *Scheduler) cleanupCooldowns(ctx context.Context) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `
UPDATE vault_credentials
SET cooldown_until = NULL
WHERE cooldown_until IS NOT NULL AND cooldown_until <= ?`, now)
	return err
}

func (s *Scheduler) pruneLogs(ctx context.Context) error {
	cutoff := time.Now().UTC().AddDate(0, 0, -s.retentionDays)
	_, _ = s.db.ExecContext(ctx, "DELETE FROM audit_logs WHERE created_at < ?", cutoff)
	_, _ = s.db.ExecContext(ctx, "DELETE FROM webhook_deliveries WHERE created_at < ?", cutoff)
	_, _ = s.db.ExecContext(ctx, "DELETE FROM tool_executions WHERE executed_at < ?", cutoff)
	return nil
}
