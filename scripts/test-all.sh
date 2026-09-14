#!/usr/bin/env bash
# EkaRouter Automated Test Runner (Bash / CI)
set -euo pipefail

echo "=================================================="
echo "   EKAROUTER FINAL RELEASE QA & TEST SUITE RUNNER  "
echo "=================================================="

echo -e "\n[1/6] Running go fmt..."
FMT_OUTPUT=$(go fmt ./...)
if [ -n "$FMT_OUTPUT" ]; then
    echo "Formatted files: $FMT_OUTPUT"
else
    echo "All files formatted cleanly."
fi

echo -e "\n[2/6] Running go vet..."
go vet ./...
echo "Static analysis passed."

echo -e "\n[3/6] Running Unit Tests..."
go test ./internal/...
echo "Unit tests passed."

echo -e "\n[4/6] Running Integration Tests..."
go test ./tests/integration/...
echo "Integration tests passed."

echo -e "\n[5/6] Running E2E Tests..."
go test ./tests/e2e/...
echo "E2E tests passed."

echo -e "\n[6/6] Verifying Binary Build..."
go build -o ekarouter ./cmd/ekarouter
echo "Build passed: ekarouter binary created."

echo -e "\n=================================================="
echo "   ALL TESTS PASSED & BUILD VERIFIED SUCCESSFULLY   "
echo "=================================================="
