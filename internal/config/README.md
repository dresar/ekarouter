# Config Package

## Purpose
Manages runtime configuration defaults, environment variable ingestion, and validation.

## Files
- `config.go`: Defines `Config` struct, default configurations, environment loading, and validation rules.
- `config_test.go`: Unit tests verifying defaults, overrides, and validation logic.

## Allowed Responsibilities
- Reading and parsing configuration from environment variables and command line options.
- Struct-level invariant validation.

## Forbidden Responsibilities
- No network operations or database calls.
