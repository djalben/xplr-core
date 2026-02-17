# Скрипт для проверки миграций БД

Write-Host "🔍 Проверка миграций БД..." -ForegroundColor Yellow
Write-Host ""

$postgresRunning = docker ps --filter "name=xplr-postgres" --format "{{.Names}}"
if (-not $postgresRunning) {
    Write-Host "❌ Контейнер xplr-postgres не запущен. Запустите: docker compose up -d" -ForegroundColor Red
    exit 1
}

# Проверка новых полей в cards
Write-Host "Проверка новых полей в таблице cards..." -ForegroundColor Cyan
$columnsQuery = "SELECT column_name FROM information_schema.columns WHERE table_name = 'cards' AND column_name IN ('auto_replenish_enabled', 'team_id', 'card_type', 'card_balance');"
$columns = docker exec xplr-postgres psql -U xplr_user -d xplr_db -t -c $columnsQuery 2>&1

if ($columns -like "*auto_replenish_enabled*") {
    Write-Host "✅ auto_replenish_enabled - найдено" -ForegroundColor Green
} else {
    Write-Host "❌ auto_replenish_enabled - НЕ найдено" -ForegroundColor Red
}

if ($columns -like "*team_id*") {
    Write-Host "✅ team_id - найдено" -ForegroundColor Green
} else {
    Write-Host "❌ team_id - НЕ найдено" -ForegroundColor Red
}

if ($columns -like "*card_type*") {
    Write-Host "✅ card_type - найдено" -ForegroundColor Green
} else {
    Write-Host "❌ card_type - НЕ найдено" -ForegroundColor Red
}

if ($columns -like "*card_balance*") {
    Write-Host "✅ card_balance - найдено" -ForegroundColor Green
} else {
    Write-Host "❌ card_balance - НЕ найдено" -ForegroundColor Red
}

Write-Host ""

# Проверка новых таблиц
Write-Host "Проверка новых таблиц..." -ForegroundColor Cyan
$tablesQuery = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('teams', 'team_members', 'user_grades', 'referrals');"
$tables = docker exec xplr-postgres psql -U xplr_user -d xplr_db -t -c $tablesQuery 2>&1

if ($tables -like "*teams*") {
    Write-Host "✅ teams - найдено" -ForegroundColor Green
} else {
    Write-Host "❌ teams - НЕ найдено" -ForegroundColor Red
}

if ($tables -like "*team_members*") {
    Write-Host "✅ team_members - найдено" -ForegroundColor Green
} else {
    Write-Host "❌ team_members - НЕ найдено" -ForegroundColor Red
}

if ($tables -like "*user_grades*") {
    Write-Host "✅ user_grades - найдено" -ForegroundColor Green
} else {
    Write-Host "❌ user_grades - НЕ найдено" -ForegroundColor Red
}

if ($tables -like "*referrals*") {
    Write-Host "✅ referrals - найдено" -ForegroundColor Green
} else {
    Write-Host "❌ referrals - НЕ найдено" -ForegroundColor Red
}

Write-Host ""
Write-Host "Если какие-то поля/таблицы не найдены, выполните миграции:" -ForegroundColor Yellow
Write-Host "  .\apply_migrations.ps1" -ForegroundColor White
Write-Host "  или: docker cp backend/schema.sql xplr-postgres:/tmp/schema.sql" -ForegroundColor White
Write-Host "       docker exec xplr-postgres psql -U xplr_user -d xplr_db -f /tmp/schema.sql" -ForegroundColor White
