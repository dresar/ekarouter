# Testing Guide

All tests in EkaRouter use pure Go and in-memory or temporary SQLite databases. No real external credentials or internet network requests are made during unit testing.

## Running Tests

### Standard Test Run
```powershell
go test ./...
```

### Verbose Package Run
```powershell
go test -v ./internal/vault ./internal/platform ./internal/limits ./internal/rotator ./internal/executor ./internal/webhooks ./internal/rbac ./internal/audit ./internal/scheduler ./internal/httpapi
```

### Race Detection Run
```powershell
go test -race ./internal/vault ./internal/limits ./internal/rotator ./internal/webhooks ./internal/audit
```

### Coverage Report
```powershell
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```
