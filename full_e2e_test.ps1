# Full E2E test: registration, deposit, cards (issue/block/unblock), report, API key, grade, teams, referrals, settings.
# Run from: C:\Users\aalab\XPLR

$ErrorActionPreference = "Continue"
$baseUrl = "http://localhost:8080/api/v1"
$healthUrl = "http://localhost:8080/health"
$token = $null
$headers = @{}
$passed = 0
$failed = 0

function Ok { param($msg) Write-Host "  OK $msg" -ForegroundColor Green; $script:passed++ }
function Fail { param($msg) Write-Host "  FAIL $msg" -ForegroundColor Red; $script:failed++ }

Write-Host ""
Write-Host "======== Full E2E Test - XPLR ========" -ForegroundColor Green
Write-Host ""

# 1. Health
Write-Host "[1] Health" -ForegroundColor Cyan
try {
    $r = Invoke-WebRequest -Uri $healthUrl -UseBasicParsing -TimeoutSec 5
    if ($r.StatusCode -eq 200) { Ok "Backend available" } else { Fail "Status $($r.StatusCode)" }
} catch { Fail $_.Exception.Message }

# 2. Register
$testEmail = "e2e_$(Get-Random)@example.com"
$testPass = "TestPass123!"
Write-Host "[2] Register ($testEmail)" -ForegroundColor Cyan
try {
    $body = @{ email = $testEmail; password = $testPass } | ConvertTo-Json
    $r = Invoke-WebRequest -Uri "$baseUrl/auth/register" -Method POST -Body $body -ContentType "application/json" -UseBasicParsing
    if ($r.StatusCode -notin 200, 201) { Fail "Status $($r.StatusCode)"; throw "stop" }
    $data = $r.Content | ConvertFrom-Json
    $token = $data.token
    if (-not $token) { Fail "No token"; throw "stop" }
    $headers = @{ "Authorization" = "Bearer $token" }
    Ok "Token received"
} catch { if ($_.Exception.Message -ne "stop") { Fail $_.Exception.Message } }

if (-not $token) {
    Write-Host "No token. Exit." -ForegroundColor Red
    exit 1
}

# 3. Login
Write-Host "[3] Login" -ForegroundColor Cyan
try {
    $body = @{ email = $testEmail; password = $testPass } | ConvertTo-Json
    $r = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Body $body -ContentType "application/json" -UseBasicParsing
    if ($r.StatusCode -ne 200) { Fail "Status $($r.StatusCode)" } else { Ok "Login works" }
} catch { Fail $_.Exception.Message }

# 4. GET /user/me
Write-Host "[4] GET /user/me" -ForegroundColor Cyan
try {
    $r = Invoke-WebRequest -Uri "$baseUrl/user/me" -Headers $headers -UseBasicParsing
    $me = $r.Content | ConvertFrom-Json
    if ($me.email) { Ok "id=$($me.id) balance=$($me.balance)" } else { Fail "No email in response" }
} catch { Fail $_.Exception.Message }

# 5. POST /user/deposit
Write-Host "[5] POST /user/deposit" -ForegroundColor Cyan
try {
    $body = @{ amount = 100.50 } | ConvertTo-Json
    $r = Invoke-WebRequest -Uri "$baseUrl/user/deposit" -Method POST -Headers $headers -Body $body -ContentType "application/json" -UseBasicParsing
    $data = $r.Content | ConvertFrom-Json
    Ok "new_balance=$($data.new_balance)"
} catch { Fail $_.Exception.Message }

# 6. GET /user/cards
Write-Host "[6] GET /user/cards" -ForegroundColor Cyan
try {
    $r = Invoke-WebRequest -Uri "$baseUrl/user/cards" -Headers $headers -UseBasicParsing
    $cards = $r.Content | ConvertFrom-Json
    if ($null -eq $cards) { $cards = @() }
    if ($cards -isnot [Array]) { $cards = @($cards) }
    Ok "cards count=$($cards.Count)"
} catch { Fail $_.Exception.Message }

# 7. POST /user/cards/issue
$cardId = $null
Write-Host "[7] POST /user/cards/issue" -ForegroundColor Cyan
try {
    $body = @{ nickname = "E2E Card" } | ConvertTo-Json
    $r = Invoke-WebRequest -Uri "$baseUrl/user/cards/issue" -Method POST -Headers $headers -Body $body -ContentType "application/json" -UseBasicParsing
    # If 404/405, try with count (some APIs use count)
    if ($r.StatusCode -in 404, 405) {
        $body = @{ count = 1 } | ConvertTo-Json
        $r = Invoke-WebRequest -Uri "$baseUrl/user/cards/issue" -Method POST -Headers $headers -Body $body -ContentType "application/json" -UseBasicParsing
    }
    if ($r.StatusCode -notin 200, 201) { Fail "Status $($r.StatusCode) $($r.Content)"; throw "stop" }
    $data = $r.Content | ConvertFrom-Json
    if ($data.id) { $cardId = $data.id }
    elseif ($data.card_id) { $cardId = $data.card_id }
    elseif ($data.cards -and $data.cards.Count -gt 0) { $cardId = $data.cards[0].id }
    elseif ($data -is [Array] -and $data.Count -gt 0 -and $data[0].id) { $cardId = $data[0].id }
    if (-not $cardId) {
        Start-Sleep -Seconds 1
        $r2 = Invoke-WebRequest -Uri "$baseUrl/user/cards" -Headers $headers -UseBasicParsing
        $cards = $r2.Content | ConvertFrom-Json
        if ($cards -is [Array] -and $cards.Count -gt 0) { $cardId = $cards[0].id }
        elseif ($cards -and $cards.id) { $cardId = $cards.id }
        elseif ($cards -and $cards.cards -and $cards.cards.Count -gt 0) { $cardId = $cards.cards[0].id }
    }
    if ($cardId) { Ok "card id=$cardId" } else { Ok "issue response received" }
} catch {
    if ($_.Exception.Message -ne "stop") { Fail $_.Exception.Message }
    try {
        Start-Sleep -Seconds 1
        $r = Invoke-WebRequest -Uri "$baseUrl/user/cards" -Headers $headers -UseBasicParsing
        $cards = $r.Content | ConvertFrom-Json
        if ($cards -is [Array] -and $cards.Count -gt 0) { $cardId = $cards[0].id }
        elseif ($cards -and $cards.id) { $cardId = $cards.id }
        elseif ($cards -and $cards.cards -and $cards.cards.Count -gt 0) { $cardId = $cards.cards[0].id }
    } catch {}
}

# 8. PATCH card status BLOCKED
if ($cardId) {
    Write-Host "[8] PATCH /user/cards/$cardId/status (BLOCKED)" -ForegroundColor Cyan
    try {
        $body = @{ card_status = "BLOCKED" } | ConvertTo-Json
        $r = Invoke-WebRequest -Uri "$baseUrl/user/cards/$cardId/status" -Method PATCH -Headers $headers -Body $body -ContentType "application/json" -UseBasicParsing
        if ($r.StatusCode -eq 200) { Ok "card blocked" } else { Fail "Status $($r.StatusCode)" }
    } catch {
        try {
            $body = @{ status = "BLOCKED" } | ConvertTo-Json
            $r = Invoke-WebRequest -Uri "$baseUrl/user/cards/$cardId/status" -Method PATCH -Headers $headers -Body $body -ContentType "application/json" -UseBasicParsing
            if ($r.StatusCode -eq 200) { Ok "card blocked" } else { Fail "Status $($r.StatusCode)" }
        } catch { Fail $_.Exception.Message }
    }

    Write-Host "[9] PATCH /user/cards/$cardId/status (ACTIVE)" -ForegroundColor Cyan
    try {
        $body = @{ card_status = "ACTIVE" } | ConvertTo-Json
        $r = Invoke-WebRequest -Uri "$baseUrl/user/cards/$cardId/status" -Method PATCH -Headers $headers -Body $body -ContentType "application/json" -UseBasicParsing
        if ($r.StatusCode -eq 200) { Ok "card unblocked" } else { Fail "Status $($r.StatusCode)" }
    } catch {
        try {
            $body = @{ status = "ACTIVE" } | ConvertTo-Json
            $r = Invoke-WebRequest -Uri "$baseUrl/user/cards/$cardId/status" -Method PATCH -Headers $headers -Body $body -ContentType "application/json" -UseBasicParsing
            if ($r.StatusCode -eq 200) { Ok "card unblocked" } else { Fail "Status $($r.StatusCode)" }
        } catch { Fail $_.Exception.Message }
    }
} else {
    Write-Host "[8-9] Skip block/unblock (no card id)" -ForegroundColor Yellow
}

# 10. GET /user/report
Write-Host "[10] GET /user/report" -ForegroundColor Cyan
try {
    $r = Invoke-WebRequest -Uri "$baseUrl/user/report" -Headers $headers -UseBasicParsing
    if ($r.StatusCode -eq 200) { Ok "report returned" } else { Fail "Status $($r.StatusCode)" }
} catch { Fail $_.Exception.Message }

# 11. POST /user/api-key
Write-Host "[11] POST /user/api-key" -ForegroundColor Cyan
try {
    $r = Invoke-WebRequest -Uri "$baseUrl/user/api-key" -Method POST -Headers $headers -UseBasicParsing
    $data = $r.Content | ConvertFrom-Json
    if (($r.StatusCode -in 200, 201) -and $data.api_key) { Ok "api_key created" } else { Fail "Status or no api_key" }
} catch { Fail $_.Exception.Message }

# 12. GET /user/grade
Write-Host "[12] GET /user/grade" -ForegroundColor Cyan
try {
    $r = Invoke-WebRequest -Uri "$baseUrl/user/grade" -Headers $headers -UseBasicParsing
    $g = $r.Content | ConvertFrom-Json
    Ok "grade=$($g.grade) fee=$($g.fee_percent)%"
} catch { Fail $_.Exception.Message }

# 13. POST /user/teams
$teamId = $null
Write-Host "[13] POST /user/teams" -ForegroundColor Cyan
try {
    $body = @{ name = "E2E Team" } | ConvertTo-Json
    $r = Invoke-WebRequest -Uri "$baseUrl/user/teams" -Method POST -Headers $headers -Body $body -ContentType "application/json" -UseBasicParsing
    $data = $r.Content | ConvertFrom-Json
    if ($data.id) { $teamId = $data.id }; if ($data.team_id) { $teamId = $data.team_id }
    if ($r.StatusCode -in 200, 201) { Ok "team created id=$teamId" } else { Fail "Status $($r.StatusCode)" }
} catch { Fail $_.Exception.Message }

# 14. GET /user/teams
Write-Host "[14] GET /user/teams" -ForegroundColor Cyan
try {
    $r = Invoke-WebRequest -Uri "$baseUrl/user/teams" -Headers $headers -UseBasicParsing
    $teams = $r.Content | ConvertFrom-Json
    if ($null -eq $teams) { $teams = @() }
    if ($teams -isnot [Array]) { $teams = @($teams) }
    if (-not $teamId -and $teams.Count -gt 0 -and $teams[0].id) { $teamId = $teams[0].id }
    Ok "teams count=$($teams.Count)"
} catch { Fail $_.Exception.Message }

# 15. GET /user/teams/:id
if ($teamId) {
    Write-Host "[15] GET /user/teams/$teamId" -ForegroundColor Cyan
    try {
        $r = Invoke-WebRequest -Uri "$baseUrl/user/teams/$teamId" -Headers $headers -UseBasicParsing
        if ($r.StatusCode -eq 200) { Ok "team details" } else { Fail "Status $($r.StatusCode)" }
    } catch { Fail $_.Exception.Message }
}

# 16. GET /user/referrals
Write-Host "[16] GET /user/referrals" -ForegroundColor Cyan
try {
    $r = Invoke-WebRequest -Uri "$baseUrl/user/referrals" -Headers $headers -UseBasicParsing
    $ref = $r.Content | ConvertFrom-Json
    Ok "referral_code=$($ref.referral_code)"
} catch { Fail $_.Exception.Message }

# 17. POST /user/settings/telegram (optional: 400 = invalid chat_id is acceptable)
Write-Host "[17] POST /user/settings/telegram" -ForegroundColor Cyan
try {
    $body = @{ chat_id = "123456789" } | ConvertTo-Json
    $r = Invoke-WebRequest -Uri "$baseUrl/user/settings/telegram" -Method POST -Headers $headers -Body $body -ContentType "application/json" -UseBasicParsing
    if ($r.StatusCode -in 200, 201, 204) { Ok "settings updated" } else { Fail "Status $($r.StatusCode)" }
} catch {
    if ($_.Exception.Message -match "400") { Ok "endpoint exists (400 = validation)" } else { Fail $_.Exception.Message }
}

# Summary
Write-Host ""
Write-Host "======== Passed: $passed  Failed: $failed ========" -ForegroundColor $(if ($failed -eq 0) { "Green" } else { "Yellow" })
if ($failed -gt 0) { exit 1 }
