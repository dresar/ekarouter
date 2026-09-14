# Contributing to EkaRouter

Thank you for your interest in contributing to EkaRouter!

EkaRouter is an enterprise-grade AI Gateway and Developer Platform with strict engineering constraints. Please review the following standards before opening a pull request.

---

## 1. Engineering Standards

### Zero Comments Rule (`/nokomen`)
- **Strict Requirement**: Under no circumstances should Go code (`*.go`) contain inline comments (`//`) or block comments (`/* */`).
- All code must be idiomatic, concise, clean, and self-documenting through precise domain nomenclature.
- Explanations, architectural rationale, and usage instructions must live in Markdown (`README.md`, `docs/`, or `.github/`).

### Pure Go & CGO-Free Architecture
- Do not introduce CGO dependencies or external C libraries.
- Database access must use pure-Go SQLite (`modernc.org/sqlite`).
- Code must compile natively across Windows, Linux, and macOS without requiring GCC/Clang.

---

## 2. Development & Testing Workflow

### Prerequisites
- Go 1.24 or higher
- Git

### Running Tests
Before submitting any changes, all tests must pass:
```powershell
go test -count=1 ./...
```

For targeted package testing:
```powershell
go test -v ./internal/gateway
go test -v ./internal/credpool
go test -v ./internal/routing
go test -v ./internal/httpapi
go test -v ./internal/rotator
```

### Build Verification
Ensure the standalone executable builds cleanly:
```powershell
go build -o bin/ekarouter.exe ./cmd/ekarouter
```

---

## 3. Git Commit Conventions

We follow the Conventional Commits specification:
- `feat(scope)`: A new feature or provider adapter
- `fix(scope)`: A bug fix or failover improvement
- `refactor(scope)`: Code refactoring without behavioral alterations
- `docs(scope)`: Documentation additions or updates
- `test(scope)`: Adding or correcting tests

Examples:
- `fix(routing): ensure deterministic tie-breaking for equal priority accounts`
- `feat(providers): add streaming SSE support for custom backend`
- `docs(identity): update project architecture manifest`

---

## 4. Submitting a Pull Request

1. Fork the repository and create your branch from `main`:
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. Commit your changes following conventional commits.
3. Run the complete test suite (`go test ./...`) and build verification.
4. Verify that no comments were added to Go source files.
5. Push your branch and open a Pull Request against `main`.
