# DB Package

## Purpose
Manages SQLite connections, pragmas (WAL, foreign keys, busy timeout), automated migration execution, and consistent zero-corruption VACUUM backups.

## Files
- `db.go`: SQLite connection factory, WAL pragma setup, sequential file-based migration runner, and VACUUM INTO backup method.
- `db_test.go`: Tests covering database opening, schema creation, migrations validation, and backup execution.

## Allowed Responsibilities
- SQLite connection lifecycle and connection pool sizing.
- Schema version tracking and migrations execution.
- Online transactional database backup.

## Forbidden Responsibilities
- No business routing, provider adapters, or HTTP logic.
