# Скрипт для полной очистки Docker

Write-Host "🧹 Полная очистка Docker для XPLR" -ForegroundColor Red
Write-Host "=======================================" -ForegroundColor Red
Write-Host ""
Write-Host "⚠️ ВНИМАНИЕ: Это удалит ВСЕ контейнеры, volumes и images проекта!" -ForegroundColor Yellow
Write-Host ""

$confirm = Read-Host "Продолжить? (y/N)"
if ($confirm -ne "y" -and $confirm -ne "Y") {
    Write-Host "Отменено." -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "1️⃣ Остановка и удаление контейнеров..." -ForegroundColor Yellow
docker compose down -v
if ($LASTEXITCODE -ne 0) {
    Write-Host "⚠️ Ошибка при остановке контейнеров" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "2️⃣ Удаление volumes проекта (уже выполнено через down -v)" -ForegroundColor Gray

Write-Host ""
Write-Host "3️⃣ Удаление images проекта (по текущему compose)..." -ForegroundColor Yellow
$imgIds = docker compose images -q 2>$null
if ($imgIds) {
    $imgIds | ForEach-Object {
        docker rmi -f $_ 2>$null
        if ($LASTEXITCODE -eq 0) { Write-Host "   Удален image: $_" -ForegroundColor Gray }
    }
} else {
    Write-Host "   ℹ️ Нет образов compose или контейнеры не собраны" -ForegroundColor Gray
}

Write-Host ""
Write-Host "4️⃣ Очистка build cache..." -ForegroundColor Yellow
docker builder prune -f

Write-Host ""
Write-Host "✅ Очистка завершена!" -ForegroundColor Green
Write-Host ""
Write-Host "📝 Следующий шаг: Пересобрать контейнеры" -ForegroundColor Cyan
Write-Host "   docker compose build --no-cache" -ForegroundColor White
Write-Host "   docker compose up -d" -ForegroundColor White
Write-Host ""
