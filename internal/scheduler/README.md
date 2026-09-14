# Scheduler Package

## Purpose
The `scheduler` package provides a background job runner for recurring platform maintenance, including credential cooldown cleanup, log retention pruning, and scheduled tasks.

## Responsibilities
- Periodic ticker loop running with context cancellation for clean, graceful shutdown.
- State and execution recording in `scheduled_tasks` and `task_runs` tables.
- Automatic cooldown reset for credentials whose `cooldown_until` has elapsed.
- Log retention policy enforcement, deleting records in `audit_logs`, `webhook_deliveries`, and `tool_executions` older than configured retention days.

## Public Interfaces
- `NewScheduler(db *sql.DB, registry *platform.Registry, retentionDays int, interval time.Duration) *Scheduler`
- `(*Scheduler) Start(ctx context.Context)`
- `(*Scheduler) Stop()`
- `(*Scheduler) ExecuteTask(ctx context.Context, taskType string, fn JobFunc) error`

## Security Considerations
- Background jobs run with bounded goroutines and do not spawn unconstrained workers.
- Tasks handle database transactions safely and report execution errors to `task_runs`.
