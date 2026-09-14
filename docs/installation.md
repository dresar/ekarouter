# Installation Guide

## System Requirements
- OS: Windows 10/11 / Windows Server, Linux (x86_64 / ARM64), or macOS.
- Architecture: 64-bit AMD64 or ARM64.
- Memory: Minimum 64MB RAM.
- Disk Space: ~30MB for binary and SQLite database.
- Dependencies: None (No Node.js, Python, or external database required).

## Building from Source

### Prerequisites
- Go 1.22 or newer installed (`go version`).

### Windows Compilation
Open PowerShell in the repository root and execute:

```powershell
go build -ldflags="-s -w" -o ekarouter.exe ./cmd/ekarouter
```

### Linux / macOS Compilation
```bash
go build -ldflags="-s -w" -o ekarouter ./cmd/ekarouter
```

## Running EkaRouter

### First Run
Copy `.env.example` to `.env`:
```powershell
Copy-Item .env.example .env
```

Start the application:
```powershell
.\ekarouter.exe serve
```

EkaRouter will automatically:
1. Initialize SQLite database at `data/ekarouter.db`.
2. Apply migrations `0001_initial.sql`, `0002_credential_pools.sql`, and `0003_developer_platform.sql`.
3. Register default built-in providers across 11 categories.
4. Bind to `http://0.0.0.0:8080`.
