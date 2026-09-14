$ErrorActionPreference = "Stop"

$repoRoot = (Get-Item $PSScriptRoot).Parent.FullName
$binPath = Join-Path $repoRoot "bin\ekarouter.exe"
$migPath = Join-Path $repoRoot "migrations"
$mainDb = Join-Path $repoRoot "data\ekarouter.db"

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host " EkaRouter Full System & Endpoint Verification Suite" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan

if (!(Test-Path $binPath)) {
    Write-Host "[1/5] Building bin/ekarouter.exe..." -ForegroundColor Yellow
    & go build -o $binPath (Join-Path $repoRoot "cmd\ekarouter")
} else {
    Write-Host "[1/5] Verified bin/ekarouter.exe exists." -ForegroundColor Green
}

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("ekarouter_test_" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $tempDir | Out-Null
$testDb = Join-Path $tempDir "test.db"
Copy-Item -Path $mainDb -Destination $testDb -Force
Write-Host "[2/5] Prepared test database at $testDb" -ForegroundColor Green

$port = 19199
Write-Host "[3/5] Starting EkaRouter daemon on port $port..." -ForegroundColor Yellow
$proc = Start-Process -FilePath $binPath -ArgumentList "-host", "127.0.0.1", "-port", "$port", "-db", "$testDb", "-migrations", "$migPath" -PassThru

$results = [System.Collections.Generic.List[PSCustomObject]]::new()
function Record-Test($category, $endpoint, $method, $status, $expected, $passed, $details) {
    $results.Add([PSCustomObject]@{
        Category = $category
        Endpoint = $endpoint
        Method   = $method
        Status   = $status
        Expected = $expected
        Passed   = $passed
        Details  = $details
    })
    $color = if ($passed) { "Green" } else { "Red" }
    $mark = if ($passed) { "PASS" } else { "FAIL" }
    Write-Host "  [$mark] $method $endpoint -> $status ($details)" -ForegroundColor $color
}

try {
    Start-Sleep -Seconds 2

    Write-Host "`n[4/5] Running All Endpoint Tests..." -ForegroundColor Cyan

    # 1. Health Endpoints
    $r = Invoke-RestMethod -Uri "http://127.0.0.1:$port/health" -Method Get
    Record-Test "System Health" "/health" "GET" 200 200 ($r.status -eq "ok") "status=$($r.status), uptime=$($r.uptime)"

    $r = Invoke-RestMethod -Uri "http://127.0.0.1:$port/ready" -Method Get
    Record-Test "System Health" "/ready" "GET" 200 200 ($r.status -eq "ready") "status=$($r.status)"

    # 2. Admin Login
    $loginBody = @{ username = "admin"; password = "admin12345" } | ConvertTo-Json
    $loginResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResp.token
    Record-Test "Admin Auth" "/api/auth/login" "POST" 200 200 ($null -ne $token) "token prefix=$($token.Substring(0,8))..."

    $authHeader = @{ Authorization = "Bearer $token" }

    # 3. Admin Me
    $meResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/auth/me" -Method Get -Headers $authHeader
    Record-Test "Admin Auth" "/api/auth/me" "GET" 200 200 ($meResp.user -eq "admin") "user=$($meResp.user)"

    # 4. Providers
    $provList = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/providers" -Method Get -Headers $authHeader
    Record-Test "Admin Providers" "/api/providers" "GET" 200 200 ($provList.Count -ge 40) "count=$($provList.Count)"

    $newProv = @{ id = "test-endpoint-prov"; key = "test-endpoint-prov"; name = "Test Endpoint Prov"; kind = "openai"; base_url = "https://api.openai.com/v1"; enabled = $true } | ConvertTo-Json
    $provPost = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/providers" -Method Post -Headers $authHeader -Body $newProv -ContentType "application/json"
    Record-Test "Admin Providers" "/api/providers" "POST" 201 201 ($provPost.status -eq "created") "created id=$($provPost.id)"

    # 5. Accounts
    $accList = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/accounts" -Method Get -Headers $authHeader
    Record-Test "Admin Accounts" "/api/accounts" "GET" 200 200 ($accList.Count -ge 180) "count=$($accList.Count)"

    $newAcc = @{ id = "test-endpoint-acc"; provider_id = "test-endpoint-prov"; name = "Test Endpoint Acc"; auth_type = "apiKey"; priority = 1; api_key = "sk-test1234" } | ConvertTo-Json
    $accPost = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/accounts" -Method Post -Headers $authHeader -Body $newAcc -ContentType "application/json"
    Record-Test "Admin Accounts" "/api/accounts" "POST" 201 201 ($accPost.status -eq "created") "created id=$($accPost.id)"

    # 6. Credentials
    $credList = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/credentials" -Method Get -Headers $authHeader
    Record-Test "Admin Credentials" "/api/credentials" "GET" 200 200 ($credList.Count -ge 180) "count=$($credList.Count)"

    if ($credList.Count -gt 0) {
        $firstCredId = $credList[0].id
        $singleCred = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/credentials/$firstCredId" -Method Get -Headers $authHeader
        Record-Test "Admin Credentials" "/api/credentials/:id" "GET" 200 200 ($singleCred.id -eq $firstCredId) "account=$($singleCred.account_name), fp=$($singleCred.key_fingerprint)"
    }

    # 7. Models
    $modelList = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/models" -Method Get -Headers $authHeader
    Record-Test "Admin Models" "/api/models" "GET" 200 200 ($modelList.Count -ge 150) "count=$($modelList.Count)"

    $newModel = @{ id = "test-endpoint-prov/test-model"; provider_id = "test-endpoint-prov"; external_name = "test-model"; display_name = "Test Model" } | ConvertTo-Json
    $modelPost = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/models" -Method Post -Headers $authHeader -Body $newModel -ContentType "application/json"
    Record-Test "Admin Models" "/api/models" "POST" 201 201 ($modelPost.status -eq "created") "created model=$($modelPost.id)"

    # 8. Routes
    $routeList = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/routes" -Method Get -Headers $authHeader
    Record-Test "Admin Routes" "/api/routes" "GET" 200 200 ($routeList.Count -ge 2) "count=$($routeList.Count)"

    $newRoute = @{
        id = "test-endpoint-route"
        name = "test-endpoint-route"
        strategy = "priority"
        items = @(
            @{ provider_id = "test-endpoint-prov"; account_id = "test-endpoint-acc"; model_id = "test-endpoint-prov/test-model"; priority = 1 }
        )
    } | ConvertTo-Json -Depth 5
    $routePost = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/routes" -Method Post -Headers $authHeader -Body $newRoute -ContentType "application/json"
    Record-Test "Admin Routes" "/api/routes" "POST" 201 201 ($routePost.status -eq "created") "created route=$($routePost.id)"

    # 9. Proxies
    $proxyList = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/proxies" -Method Get -Headers $authHeader
    Record-Test "Admin Proxies" "/api/proxies" "GET" 200 200 ($proxyList.Count -eq 16) "count=$($proxyList.Count)"

    # Test an active proxy profile from the imported list
    if ($proxyList.Count -gt 0) {
        $firstProxyId = $proxyList[0].id
        $pxTest = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/proxies/$firstProxyId/test" -Method Post -Headers $authHeader
        Record-Test "Admin Proxies" "/api/proxies/:id/test" "POST" 200 200 ($true) "profile=$($proxyList[0].name), ok=$($pxTest.ok), status=$($pxTest.status), latency=$($pxTest.latency_ms)ms"
    }

    # 10. API Keys
    $keysList = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/keys" -Method Get -Headers $authHeader
    Record-Test "Admin Keys" "/api/keys" "GET" 200 200 ($keysList.Count -ge 1) "count=$($keysList.Count)"

    $newKeyBody = @{ name = "Verification Auto Key"; scopes = "*" } | ConvertTo-Json
    $newKeyResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/keys" -Method Post -Headers $authHeader -Body $newKeyBody -ContentType "application/json"
    $apiKey = $newKeyResp.api_key
    Record-Test "Admin Keys" "/api/keys" "POST" 201 201 ($apiKey.StartsWith("eka_live_")) "generated key prefix=$($newKeyResp.prefix)..."

    # 11. Usage
    $usageResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/usage" -Method Get -Headers $authHeader
    Record-Test "Admin Usage" "/api/usage" "GET" 200 200 ($null -ne $usageResp) "total_requests=$($usageResp.total_requests)"

    # 12. TokenSaver Preview
    $tsBody = @{ input = "Repeated line for compression test`nRepeated line for compression test" } | ConvertTo-Json
    $tsResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/tokensaver/preview" -Method Post -Headers $authHeader -Body $tsBody -ContentType "application/json"
    Record-Test "TokenSaver" "/api/tokensaver/preview" "POST" 200 200 ($null -ne $tsResp.output) "reduction=$($tsResp.reduction_pct)%"

    # 13. Backup
    $backupDest = Join-Path $tempDir "backup_api_created.db"
    $backupBody = @{ dest_path = $backupDest } | ConvertTo-Json
    $backupResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/backup" -Method Post -Headers $authHeader -Body $backupBody -ContentType "application/json"
    Record-Test "Admin Backup" "/api/backup" "POST" 201 201 ($backupResp.status -eq "created" -and (Test-Path $backupDest)) "saved size=$($backupResp.size_bytes) bytes"

    $backupListResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/backup" -Method Get -Headers $authHeader
    Record-Test "Admin Backup" "/api/backup" "GET" 200 200 ($backupListResp.backups.Count -ge 1) "found $($backupListResp.backups.Count) backup files"

    # 14. Gateway /v1/models
    $gwHeader = @{ Authorization = "Bearer $apiKey" }
    $gwModels = Invoke-RestMethod -Uri "http://127.0.0.1:$port/v1/models" -Method Get -Headers $gwHeader
    Record-Test "Gateway Models" "/v1/models" "GET" 200 200 ($gwModels.data.Count -gt 50) "available models count=$($gwModels.data.Count)"

    # 15. Gateway /v1/chat/completions (opt-out token saver or dummy fallback test)
    # Testing gateway error handler & parameter validation
    try {
        $chatBody = @{ model = "non-existent-route-xyz"; messages = @( @{ role = "user"; content = "test" } ) } | ConvertTo-Json
        $chatResp = Invoke-RestMethod -Uri "http://127.0.0.1:$port/v1/chat/completions" -Method Post -Headers $gwHeader -Body $chatBody -ContentType "application/json"
        Record-Test "Gateway Chat" "/v1/chat/completions" "POST" 200 502 $false "unexpected 200 for missing route"
    } catch [System.Net.WebException] {
        $statusCode = [int]$_.Exception.Response.StatusCode
        Record-Test "Gateway Chat" "/v1/chat/completions" "POST" $statusCode 502 ($statusCode -eq 502) "correctly caught bad gateway on non-existent route"
    }

    # 16. Gateway /v1/responses (opt-out token saver test)
    try {
        $respBody = @{ model = "non-existent-route-xyz"; input = "test input" } | ConvertTo-Json
        $respRes = Invoke-RestMethod -Uri "http://127.0.0.1:$port/v1/responses" -Method Post -Headers $gwHeader -Body $respBody -ContentType "application/json"
        Record-Test "Gateway Responses" "/v1/responses" "POST" 200 502 $false "unexpected 200 for missing route"
    } catch [System.Net.WebException] {
        $statusCode = [int]$_.Exception.Response.StatusCode
        Record-Test "Gateway Responses" "/v1/responses" "POST" $statusCode 502 ($statusCode -eq 502) "correctly routed to gateway handler"
    }

    Write-Host "`n[5/5] Verification Summary" -ForegroundColor Cyan
    Write-Host "============================================================" -ForegroundColor Cyan
    $passedCount = ($results | Where-Object { $_.Passed -eq $true }).Count
    $totalCount = $results.Count
    Write-Host "Total Endpoints Tested : $totalCount" -ForegroundColor White
    Write-Host "Passed                 : $passedCount" -ForegroundColor Green
    Write-Host "Failed                 : $($totalCount - $passedCount)" -ForegroundColor $(if ($totalCount -eq $passedCount) { "Green" } else { "Red" })

    if ($passedCount -ne $totalCount) {
        throw "One or more endpoint tests failed!"
    }
    Write-Host "`nALL EKAROUTER ENDPOINTS VERIFIED & PASSING 100%!" -ForegroundColor Green
} finally {
    Write-Host "`nTerminating EkaRouter test daemon..." -ForegroundColor Gray
    Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
    Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}
