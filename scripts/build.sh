#!/usr/bin/env bash
set -euo pipefail

BIN_DIR="$(dirname "$0")/../bin"
mkdir -p "$BIN_DIR"

VERSION="1.0.0"
LDFLAGS="-s -w -X main.Version=${VERSION}"

echo "Building Windows amd64 binary..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "${BIN_DIR}/ekarouter.exe" ./cmd/ekarouter

echo "Building Linux amd64 binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "${BIN_DIR}/ekarouter-linux-amd64" ./cmd/ekarouter

echo "Building Linux arm64 binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o "${BIN_DIR}/ekarouter-linux-arm64" ./cmd/ekarouter

echo "Calculating checksums..."
cd "${BIN_DIR}"
sha256sum ekarouter.exe ekarouter-linux-amd64 ekarouter-linux-arm64 > checksums.txt
cat checksums.txt

echo "Build complete."
