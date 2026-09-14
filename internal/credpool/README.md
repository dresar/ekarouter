# Credential Pool Engine (`internal/credpool`)

## 1. Purpose
The `credpool` package manages dynamic multi-credential pools, account tiering, failure tracking, and automated failover routing for API endpoints.

## 2. Responsibilities
- Credential pool creation and membership lifecycle.
- Failover policy execution (priority, round-robin, least-used, lowest-latency).
- Health check dispatch and cooldown state management.
- Failure classification and automatic quarantine of misbehaving credentials.

## 3. Public Interfaces
- `Engine`: Manages in-memory pool state, key rotation, and selection.
- `Store`: SQLite persistence layer for pools, members, and health records.
- `HealthChecker`: Performs background connectivity probes.

## 4. Dependencies
- Standard library: `context`, `database/sql`, `sync`, `time`.
- `github.com/dresar/ekarouter/internal/db`: SQLite database handle.

## 5. Security Considerations
- Credentials within pools are referenced by ID. Decrypted values are resolved only at the execution boundary.
- Concurrency-safe state machines prevent duplicate selection under high load.

## 6. Testing Instructions
Run credpool tests:
```bash
go test -v ./internal/credpool/...
```

## 7. Extension Instructions
To implement a new selection strategy, define the policy in `engine.go` and implement the ordering algorithm in `SelectCredential`.
