.PHONY: all build build-windows build-linux build-macos test test-unit test-integration test-e2e test-race test-live test-security test-api test-providers test-rotation test-proxy test-models docs openapi lint vet fmt clean migrate run

all: fmt vet test build

build:
	go build -o ekarouter ./cmd/ekarouter

build-windows:
	go build -o ekarouter.exe ./cmd/ekarouter

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/ekarouter-linux-amd64 ./cmd/ekarouter

build-macos:
	GOOS=darwin GOARCH=arm64 go build -o bin/ekarouter-darwin-arm64 ./cmd/ekarouter

run:
	go run ./cmd/ekarouter serve

migrate:
	go run ./cmd/ekarouter serve -port 0 -host 127.0.0.1

test:
	go test -v ./...

test-unit:
	go test -v ./internal/...

test-integration:
	go test -v ./tests/integration/...

test-e2e:
	go test -v ./tests/e2e/...

test-race:
	@echo "Note: -race requires CGO/gcc which is not present by default on pure Windows"
	go test -race ./internal/... || true

test-live:
	EKAROUTER_ENABLE_LIVE_PROVIDER_TESTS=true go test -v ./tests/integration/... -run TestLive

test-security:
	go test -v ./tests/integration -run TestSecurity

test-api:
	go test -v ./tests/integration -run TestAllEndpoints

test-providers:
	go test -v ./internal/providers/...

test-rotation:
	go test -v ./tests/integration -run TestRotator

test-proxy:
	go test -v ./tests/integration -run TestProxy

test-models:
	go test -v ./tests/integration -run TestGatewayRoutes

docs:
	@echo "Documentation available in docs/final-qa, docs/api, and docs/frontend"

openapi:
	@echo "OpenAPI 3.1 specification available in docs/api/openapi.yaml and docs/api/openapi.json"

lint:
	go vet ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin/ data/*.db data/*.db-wal data/*.db-shm ekarouter ekarouter.exe
