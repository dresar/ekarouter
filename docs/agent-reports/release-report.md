# Release and Build Report: EkaRouter v1.0.0

## 1. Release Overview

EkaRouter v1.0.0 provides a single, independently deployable native executable with zero external runtime dependencies (no Node.js, Python, or Go required on target machines).

## 2. Target Matrix & Artifacts

All binaries were cross-compiled with stripped symbols (`-ldflags "-s -w -X main.Version=1.0.0"`):

| Target Artifact | Platform / Arch | Size | SHA-256 Checksum |
| :--- | :--- | :--- | :--- |
| `bin/ekarouter.exe` | Windows amd64 | ~12.08 MB | `27a7cd5dd6d1e7fe303ebd84b35f9b209ff22700abf355346e3770500f5293bc` |
| `bin/ekarouter-linux-amd64` | Linux amd64 | ~11.79 MB | `c4247535c473124aeb2b9e499dd5da1eaac9d50f0dbf1816ab16a665f10a4c5e` |
| `bin/ekarouter-linux-arm64` | Linux arm64 | ~11.19 MB | `bbd4098d95d144e22b062dd190224a8f04230e7467982a719297f6ecd3b06e7a` |

## 3. Verification & Smoke Test
- **Clean Directory Smoke Test (`scripts/smoke_test.ps1`)**:
  - Successfully spawned `ekarouter.exe` in an isolated temporary directory (`AppData/Local/Temp`).
  - Validated `/health` (status `ok`, uptime < 1s).
  - Validated `/ready` (status `ready`).
  - Executed admin login and obtained session token.
  - Created an API key via `/api/keys`.
  - Queried `/v1/models` using the newly generated API key.
  - Executed live database backup command (`ekarouter.exe -backup ...`) and verified output file.
  - Terminated process cleanly.
