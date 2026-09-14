# Build & Compilation Guide

EkaRouter is engineered to compile into a single standalone executable without any CGO dependency.

## Compiling Windows EXE

```powershell
go build -ldflags="-s -w" -o ekarouter.exe ./cmd/ekarouter
```

- Output binary: `ekarouter.exe`
- Size: ~20MB
- Requires: No external DLLs, no MSVC runtime.

## Cross-Compiling for Linux

```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -ldflags="-s -w" -o bin/ekarouter-linux-amd64 ./cmd/ekarouter
```

## Cross-Compiling for macOS (Darwin)

```powershell
$env:GOOS="darwin"; $env:GOARCH="arm64"; $env:CGO_ENABLED="0"
go build -ldflags="-s -w" -o bin/ekarouter-darwin-arm64 ./cmd/ekarouter
```
