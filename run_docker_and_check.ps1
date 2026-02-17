# ============================================================
# Запуск сервиса в Docker Desktop и проверка работы
# Запускать ТОЛЬКО из основной папки: C:\Users\aalab\epn-killer-project
# ============================================================

$MainFolder = "C:\Users\aalab\epn-killer-project"
if ($PSScriptRoot -ne $MainFolder) {
    Write-Host "Запускайте этот скрипт из основной папки: $MainFolder"
    Write-Host "Текущая папка: $PSScriptRoot"
    exit 1
}

Write-Host "Папка проекта: $PSScriptRoot"
Write-Host ""

# .env
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "Создан .env из .env.example. При необходимости отредактируйте .env"
}
Write-Host "Запуск контейнеров: docker compose up -d --build"
docker compose up -d --build
if ($LASTEXITCODE -ne 0) {
    Write-Host "Ошибка запуска. Убедитесь, что Docker Desktop запущен."
    exit 1
}

Write-Host ""
Write-Host "Ожидание готовности backend (15 сек)..."
Start-Sleep -Seconds 15

# Проверка health
$healthUrl = "http://localhost:8080/health"
Write-Host "Проверка health: $healthUrl"
try {
    $r = Invoke-WebRequest -Uri $healthUrl -UseBasicParsing -TimeoutSec 5
    Write-Host "Health OK: $($r.StatusCode) - $($r.Content)"
} catch {
    Write-Host "Health не ответил: $_"
    Write-Host "Логи backend: docker compose logs backend"
    exit 1
}

Write-Host ""
Write-Host "Сервис запущен."
Write-Host "  Frontend:  http://localhost"
Write-Host "  Backend:   http://localhost:8080"
Write-Host "  Health:    http://localhost:8080/health"
Write-Host "Контейнеры: docker compose ps"
