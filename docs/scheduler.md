# Background Scheduler & Maintenance

The EkaRouter background scheduler runs periodic tasks without blocking user requests or requiring external cron tools.

## Recurring Tasks

1. **Cooldown Cleanup (`cleanup_cooldowns`):**
   Scans `vault_credentials` for records where `cooldown_until <= now` and clears the cooldown lock.

2. **Log Retention Pruning (`prune_logs`):**
   Deletes records in `audit_logs`, `webhook_deliveries`, and `tool_executions` older than `EKAROUTER_LOG_RETENTION_DAYS` (default 30 days).

3. **Task Tracking:**
   Every execution is logged into `scheduled_tasks` and `task_runs` with start time, completion time, and error details.

4. **Graceful Shutdown:**
   The scheduler loop honors `context.Context` cancellation and halts immediately upon server shutdown.
