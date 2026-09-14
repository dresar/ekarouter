# Scripts Directory

## Purpose
Build automation, cross-compilation, end-to-end smoke testing, and microbenchmarks for hot-path operations.

## Files
- `build.ps1`: Cross-compiles stripped CGO-free binaries for Windows amd64, Linux amd64, and Linux arm64 with SHA-256 checksums.
- `build.sh`: Bash counterpart for POSIX environments.
- `smoke_test.ps1`: End-to-end integration smoke test running from an isolated clean temporary directory.
- `bench_test.go`: Performance benchmarks measuring token saver latency, token hashing, and route target selection.

## Allowed Responsibilities
- Compilation automation, automated validation scripts, and performance profiling.

## Forbidden Responsibilities
- No production source code.
