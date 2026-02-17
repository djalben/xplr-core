# PowerShell скрипт для развертывания и тестирования

Write-Host "🚀 Развертывание EPN KILLER" -ForegroundColor Green
Write-Host "============================" -ForegroundColor Green
Write-Host ""

# Проверка Docker
Write-Host "1️⃣ Проверка Docker..." -ForegroundColor Yellow
$dockerCheck = $false
$maxDockerRetries = 5
$dockerRetryCount = 0

while ($dockerRetryCount -lt $maxDockerRetries -and -not $dockerCheck) {
    try {
        $dockerOutput = docker version 2>&1
        if ($LASTEXITCODE -eq 0) {
            $dockerCheck = $true
            Write-Host "✅ Docker доступен" -ForegroundColor Green
        } else {
            throw "Docker не отвечает"
        }
    } catch {
        $dockerRetryCount++
        if ($dockerRetryCount -lt $maxDockerRetries) {
            Write-Host "   Попытка подключения к Docker: $dockerRetryCount/$maxDockerRetries..." -ForegroundColor Yellow
            Write-Host "   Убедитесь, что Docker Desktop запущен и полностью загружен" -ForegroundColor Yellow
            Start-Sleep -Seconds 3
        } else {
            Write-Host "❌ Docker недоступен!" -ForegroundColor Red
            Write-Host "   Пожалуйста:" -ForegroundColor Yellow
            Write-Host "   1. Запустите Docker Desktop" -ForegroundColor White
            Write-Host "   2. Дождитесь полной загрузки (зеленая иконка в трее)" -ForegroundColor White
            Write-Host "   3. Запустите скрипт снова" -ForegroundColor White
            Write-Host ""
            Write-Host "   Или запустите PowerShell от имени администратора" -ForegroundColor Yellow
            exit 1
        }
    }
}

# Остановка и удаление контейнеров
Write-Host ""
Write-Host "2️⃣ Остановка существующих контейнеров..." -ForegroundColor Yellow
docker compose down -v
Write-Host "✅ Контейнеры остановлены" -ForegroundColor Green

# Пересборка
Write-Host ""
Write-Host "3️⃣ Пересборка контейнеров (это может занять несколько минут)..." -ForegroundColor Yellow
docker compose build --no-cache
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Ошибка при сборке!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Контейнеры пересобраны" -ForegroundColor Green

# Запуск
Write-Host ""
Write-Host "4️⃣ Запуск контейнеров..." -ForegroundColor Yellow
docker compose up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Ошибка при запуске!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Контейнеры запущены" -ForegroundColor Green

# Ожидание готовности
Write-Host ""
Write-Host "5️⃣ Ожидание готовности сервисов..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Проверка health check
Write-Host ""
Write-Host "6️⃣ Проверка health check..." -ForegroundColor Yellow
$maxRetries = 30
$retryCount = 0
$healthOk = $false

while ($retryCount -lt $maxRetries -and -not $healthOk) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 5
        if ($response.StatusCode -eq 200) {
            $healthOk = $true
            Write-Host "✅ Backend доступен!" -ForegroundColor Green
        }
    } catch {
        $retryCount++
        Write-Host "   Попытка $retryCount/$maxRetries..." -ForegroundColor Gray
        Start-Sleep -Seconds 2
    }
}

if (-not $healthOk) {
    Write-Host "⚠️ Backend не отвечает, но контейнеры запущены" -ForegroundColor Yellow
    Write-Host "   Проверьте логи: docker compose logs backend" -ForegroundColor Yellow
}

# Статус контейнеров
Write-Host ""
Write-Host "7️⃣ Статус контейнеров:" -ForegroundColor Yellow
docker compose ps

Write-Host ""
Write-Host "============================" -ForegroundColor Green
Write-Host "✅ Развертывание завершено!" -ForegroundColor Green
Write-Host ""
Write-Host "📝 Следующие шаги:" -ForegroundColor Cyan
Write-Host "   1. Откройте браузер: http://localhost" -ForegroundColor White
Write-Host "   2. Зарегистрируйте пользователя" -ForegroundColor White
Write-Host "   3. Проверьте все функции:" -ForegroundColor White
Write-Host "      - Grade System (должен быть STANDARD 6.7%)" -ForegroundColor White
Write-Host "      - Автопополнение карт" -ForegroundColor White
Write-Host "      - Команды (/teams)" -ForegroundColor White
Write-Host "      - Фильтры транзакций" -ForegroundColor White
Write-Host "      - Реферальная программа (/referrals)" -ForegroundColor White
Write-Host ""
Write-Host "📊 Просмотр логов:" -ForegroundColor Cyan
Write-Host "   docker compose logs -f" -ForegroundColor White
Write-Host ""
