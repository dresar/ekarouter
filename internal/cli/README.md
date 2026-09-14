# CLI Package (`internal/cli`)

## 1. Purpose
The `cli` package provides interactive terminal user interfaces, terminal control, and command-line management for the EkaRouter platform and AI gateway.

## 2. Responsibilities
- Terminal ANSI coloring and branded ASCII banners.
- Cross-platform raw terminal mode and window sizing (Windows Console API and POSIX VT100 fallback).
- Interactive curses-like menus for real-time model routing, provider health inspection, and credential switching.
- Gateway process orchestration and status probing.

## 3. Public Interfaces
- `RunMenu(ctx context.Context, cfg *config.Config)`: Launches the interactive dashboard terminal session.
- `RenderBanner()`: Outputs the stylized EkaRouter logo and startup metadata.
- `CheckGatewayStatus(port int) bool`: Probes the running local gateway HTTP instance.

## 4. Dependencies
- Standard library: `context`, `fmt`, `os`, `syscall`, `unsafe` (on Windows for `kernel32.dll` console handles).
- `github.com/dresar/ekarouter/internal/config`: System configuration definitions.

## 5. Security Considerations
- Zero sensitive data exposure: Secrets and API keys are strictly masked before printing to stdout/stderr.
- Escape sequence sanitization prevents terminal injection attacks.

## 6. Testing Instructions
Run CLI package unit tests:
```bash
go test -v ./internal/cli/...
```

## 7. Extension Instructions
To add a new interactive management command, add the handler to `commands.go` and register the menu key in `menu.go`.
