$ErrorActionPreference = "Stop"

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $tempDir | Out-Null
Write-Host "Isolated Smoke Test Directory: $tempDir"

$repoRoot = (Get-Item $PSScriptRoot).Parent.FullName
$binPath = Join-Path $repoRoot "bin\ekarouter.exe"
$migPath = Join-Path $repoRoot "migrations"

if (!(Test-Path $binPath)) {
    Write-Host "Building ekarouter.exe first..."
    & go build -o $binPath (Join-Path $repoRoot "cmd\ekarouter")
}

$port = 19188
$dbPath = Join-Path $tempDir "smoke.db"

Write-Host "Starting EkaRouter process on port $port..."
$proc = Start-Process -FilePath $binPath -ArgumentList "-host", "127.0.0.1", "-port", "$port", "-db", "$dbPath", "-migrations", "$migPath" -PassThru

try {
    # Wait for process initialization
    Start-Sleep -Seconds 1

    Write-Host "Checking /health endpoint..."
    $healthResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/health" -Method Get
    if ($healthResp.status -ne "ok") {
        throw "Unexpected health status: $($healthResp.status)"
    }
    Write-Host "  Health status: $($healthResp.status), uptime: $($healthResp.uptime)"

    Write-Host "Checking /ready endpoint..."
    $readyResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/ready" -Method Get
    if ($readyResp.status -ne "ready") {
        throw "Unexpected ready status: $($readyResp.status)"
    }
    Write-Host "  Ready status: $($readyResp.status)"

    Write-Host "Logging into Admin API..."
    $loginBody = @{ username = "admin"; password = "admin12345" } | ConvertTo-Json
    $loginResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResp.token
    if (!$token) {
        throw "Login failed to return session token"
    }
    Write-Host "  Admin login successful."

    Write-Host "Creating an API Key via Admin API..."
    $keyBody = @{ name = "Smoke Test Key"; scopes = "*" } | ConvertTo-Json
    $keyResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/keys" -Method Post -Headers @{ Authorization = "Bearer $token" } -Body $keyBody -ContentType "application/json"
    $apiKey = $keyResp.api_key
    Write-Host "  API Key generated: $($keyResp.prefix)..."

    Write-Host "Checking /v1/models with generated API key..."
    $modelsResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/v1/models" -Method Get -Headers @{ Authorization = "Bearer $apiKey" }
    Write-Host "  Models endpoint returned object: $($modelsResp.object)"

    Write-Host "Testing database backup action..."
    $backupPath = Join-Path $tempDir "backup.db"
    & $binPath -db $dbPath -backup $backupPath
    if (!(Test-Path $backupPath)) {
        throw "Backup file not found at $backupPath"
    }
    Write-Host "  Database backup verified successfully."

    Write-Host "ALL SMOKE TESTS PASSED CLEANLY!"
} finally {
    Write-Host "Stopping EkaRouter process..."
    Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
    Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}
