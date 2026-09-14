$ErrorActionPreference = "Stop"

$binDir = Join-Path $PSScriptRoot "..\bin"
if (!(Test-Path $binDir)) {
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
}

$version = "1.0.0"
$ldflags = "-s -w -X main.Version=$version"

Write-Host "Building Windows amd64 binary..."
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags $ldflags -o (Join-Path $binDir "ekarouter.exe") ./cmd/ekarouter

Write-Host "Building Linux amd64 binary..."
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags $ldflags -o (Join-Path $binDir "ekarouter-linux-amd64") ./cmd/ekarouter

Write-Host "Building Linux arm64 binary..."
$env:GOOS = "linux"
$env:GOARCH = "arm64"
go build -ldflags $ldflags -o (Join-Path $binDir "ekarouter-linux-arm64") ./cmd/ekarouter

# Restore host settings
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Write-Host "Calculating SHA256 checksums..."
Get-ChildItem -Path $binDir -File | ForEach-Object {
    $hash = (Get-FileHash -Path $_.FullName -Algorithm SHA256).Hash.ToLower()
    "$hash  $($_.Name)" | Out-File -Append -FilePath (Join-Path $binDir "checksums.txt") -Encoding ascii
    Write-Host "$($_.Name): $([math]::Round($_.Length / 1MB, 2)) MB | $hash"
}

Write-Host "All builds completed successfully."
