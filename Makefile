.PHONY: all build test vet fmt clean smoke bench

all: fmt vet test build

build:
	powershell -ExecutionPolicy Bypass -File ./scripts/build.ps1

test:
	go test -v ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

bench:
	go test -bench="." -benchmem ./scripts

smoke:
	powershell -ExecutionPolicy Bypass -File ./scripts/smoke_test.ps1

clean:
	rm -rf bin/ data/*.db data/*.db-wal data/*.db-shm
