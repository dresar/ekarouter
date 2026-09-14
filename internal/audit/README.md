# Audit Package

## Purpose
The `audit` package provides asynchronous, non-blocking audit logging for sensitive operations including authentication, credential creation/rotation, tool execution, and security permission changes.

## Responsibilities
- Non-blocking worker queue (buffered channel) preventing database writes from adding latency to user requests.
- Persistent audit records stored in `audit_logs` table.
- Strict prohibition of secret values in audit records.
- Query interface with filtering by action, actor, resource type, and pagination limits.

## Public Interfaces
- `NewLogger(db *sql.DB, bufferSize int) *Logger`
- `(*Logger) Log(r *Record)`
- `(*Logger) Query(ctx context.Context, action, actorID, resourceType string, limit int) ([]*Record, error)`
- `(*Logger) Close()`

## Security Considerations
- Plaintext API keys, passwords, and tokens are never included in audit payloads.
- Audit records are immutable once persisted.
