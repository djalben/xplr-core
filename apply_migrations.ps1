# Скрипт для применения миграций БД

Write-Host "📦 Применение миграций БД..." -ForegroundColor Green
Write-Host ""

# Проверка, что контейнер запущен
$postgresRunning = docker ps --filter "name=epn-killer-postgres" --format "{{.Names}}"
if (-not $postgresRunning) {
    Write-Host "❌ Контейнер PostgreSQL не запущен!" -ForegroundColor Red
    Write-Host "   Запустите: docker compose up -d" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ Контейнер PostgreSQL запущен" -ForegroundColor Green
Write-Host ""

# Ожидание готовности БД
Write-Host "⏳ Ожидание готовности БД..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Копирование schema.sql в контейнер
Write-Host "📋 Копирование schema.sql в контейнер..." -ForegroundColor Yellow
docker cp backend/schema.sql epn-killer-postgres:/tmp/schema.sql

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Ошибка копирования файла!" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Файл скопирован" -ForegroundColor Green
Write-Host ""

# Применение миграций
Write-Host "🔧 Применение миграций..." -ForegroundColor Yellow
docker exec epn-killer-postgres psql -U epnkiller_user -d epnkiller_db -f /tmp/schema.sql

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Миграции применены успешно!" -ForegroundColor Green
} else {
    Write-Host "⚠️ Возможны ошибки при применении миграций" -ForegroundColor Yellow
    Write-Host "   Проверьте логи выше" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "📊 Проверка примененных миграций..." -ForegroundColor Cyan

# Проверка новых таблиц
$tables = docker exec epn-killer-postgres psql -U epnkiller_user -d epnkiller_db -t -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('teams', 'team_members', 'user_grades', 'referrals');"

if ($tables -like "*teams*") {
    Write-Host "✅ Таблица teams создана" -ForegroundColor Green
}
if ($tables -like "*team_members*") {
    Write-Host "✅ Таблица team_members создана" -ForegroundColor Green
}
if ($tables -like "*user_grades*") {
    Write-Host "✅ Таблица user_grades создана" -ForegroundColor Green
}
if ($tables -like "*referrals*") {
    Write-Host "✅ Таблица referrals создана" -ForegroundColor Green
}

Write-Host ""
Write-Host "✅ Готово! Миграции применены." -ForegroundColor Green
Write-Host ""
Write-Host "📝 Следующий шаг: Перезапустить backend для применения изменений" -ForegroundColor Cyan
Write-Host "   docker compose restart backend" -ForegroundColor White
