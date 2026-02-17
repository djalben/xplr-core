# Test API XPLR
# Encoding: UTF-8, use only ASCII double-quotes in strings

Write-Host "Test API XPLR" -ForegroundColor Green
Write-Host "==================" -ForegroundColor Green
Write-Host ""

$baseUrl = "http://localhost:8080/api/v1"
$healthUrl = "http://localhost:8080/health"

# 1. Health check
Write-Host "1. Health check..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri $healthUrl -UseBasicParsing -TimeoutSec 5
    if ($response.StatusCode -eq 200) {
        Write-Host "OK Backend available" -ForegroundColor Green
    }
} catch {
    Write-Host "FAIL Backend unavailable: $_" -ForegroundColor Red
}
Write-Host ""

# 2. Register
Write-Host "2. Register..." -ForegroundColor Yellow
$randomEmail = "test_$(Get-Random)@example.com"
$registerBody = @{ email = $randomEmail; password = "test123456" } | ConvertTo-Json
$token = $null
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/register" -Method POST -Body $registerBody -ContentType "application/json" -UseBasicParsing
    if ($response.StatusCode -eq 200 -or $response.StatusCode -eq 201) {
        $token = ($response.Content | ConvertFrom-Json).token
        Write-Host "OK Registered: $randomEmail" -ForegroundColor Green
    }
} catch {
    Write-Host "WARN Register: $_" -ForegroundColor Yellow
}
Write-Host ""

# 3-6. Protected endpoints
if ($token) {
    $headers = @{ "Authorization" = "Bearer $token" }

    Write-Host "3. GET /user/me..." -ForegroundColor Yellow
    try {
        $response = Invoke-WebRequest -Uri "$baseUrl/user/me" -Headers $headers -UseBasicParsing
        $userData = $response.Content | ConvertFrom-Json
        Write-Host "OK /user/me email=$($userData.email) balance=$($userData.balance)" -ForegroundColor Green
    } catch { Write-Host "FAIL /user/me: $_" -ForegroundColor Red }
    Write-Host ""

    Write-Host "4. GET /user/grade..." -ForegroundColor Yellow
    try {
        $response = Invoke-WebRequest -Uri "$baseUrl/user/grade" -Headers $headers -UseBasicParsing
        $g = $response.Content | ConvertFrom-Json
        Write-Host "OK /user/grade grade=$($g.grade) fee=$($g.fee_percent)%" -ForegroundColor Green
    } catch { Write-Host "FAIL /user/grade: $_" -ForegroundColor Red }
    Write-Host ""

    Write-Host "5. GET /user/teams..." -ForegroundColor Yellow
    try {
        $response = Invoke-WebRequest -Uri "$baseUrl/user/teams" -Headers $headers -UseBasicParsing
        Write-Host "OK /user/teams" -ForegroundColor Green
    } catch { Write-Host "FAIL /user/teams: $_" -ForegroundColor Red }
    Write-Host ""

    Write-Host "6. GET /user/referrals..." -ForegroundColor Yellow
    try {
        $response = Invoke-WebRequest -Uri "$baseUrl/user/referrals" -Headers $headers -UseBasicParsing
        Write-Host "OK /user/referrals" -ForegroundColor Green
    } catch { Write-Host "FAIL /user/referrals: $_" -ForegroundColor Red }
    Write-Host ""
}

Write-Host "==================" -ForegroundColor Green
Write-Host "Done. If errors: docker compose logs backend" -ForegroundColor Cyan
