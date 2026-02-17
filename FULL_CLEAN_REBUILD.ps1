# Полная очистка и пересборка EPN KILLER

Write-Host "🚀 Полная очистка и пересборка EPN KILLER" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
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
            Start-Sleep -Seconds 3
        } else {
            Write-Host "❌ Docker недоступен!" -ForegroundColor Red
            Write-Host "   Запустите Docker Desktop и попробуйте снова" -ForegroundColor Yellow
            exit 1
        }
    }
}

Write-Host ""
Write-Host "2️⃣ Остановка и удаление контейнеров..." -ForegroundColor Yellow
docker compose down -v
Write-Host "✅ Контейнеры остановлены" -ForegroundColor Green

Write-Host ""
Write-Host "3️⃣ Удаление volumes..." -ForegroundColor Yellow
docker volume ls -q --filter "name=epn-killer" | ForEach-Object {
    docker volume rm $_ 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✅ Удален volume: $_" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "4️⃣ Удаление images проекта..." -ForegroundColor Yellow
$images = docker images --filter "reference=epn-killer*" -q
if ($images) {
    $images | ForEach-Object {
        docker rmi -f $_ 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "   ✅ Удален image: $_" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "   ℹ️ Images не найдены" -ForegroundColor Gray
}

Write-Host ""
Write-Host "5️⃣ Очистка build cache..." -ForegroundColor Yellow
docker builder prune -f | Out-Null
Write-Host "✅ Build cache очищен" -ForegroundColor Green

Write-Host ""
Write-Host "6️⃣ Пересборка контейнеров (это может занять несколько минут)..." -ForegroundColor Yellow
docker compose build --no-cache
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Ошибка при сборке!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Контейнеры пересобраны" -ForegroundColor Green

Write-Host ""
Write-Host "7️⃣ Запуск контейнеров..." -ForegroundColor Yellow
docker compose up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Ошибка при запуске!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Контейнеры запущены" -ForegroundColor Green

Write-Host ""
Write-Host "8️⃣ Ожидание готовности сервисов..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

Write-Host ""
Write-Host "9️⃣ Проверка health check..." -ForegroundColor Yellow
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

Write-Host ""
Write-Host "🔟 Статус контейнеров:" -ForegroundColor Yellow
docker compose ps

Write-Host ""
Write-Host "============================" -ForegroundColor Green
Write-Host "✅ Очистка и пересборка завершены!" -ForegroundColor Green
Write-Host ""
Write-Host "📝 Следующие шаги:" -ForegroundColor Cyan
Write-Host "   1. Откройте браузер: http://localhost" -ForegroundColor White
Write-Host "   2. Зарегистрируйте пользователя" -ForegroundColor White
Write-Host "   3. Проверьте все функции:" -ForegroundColor White
Write-Host "      - Grade System (STANDARD 6.7%)" -ForegroundColor White
Write-Host "      - Автопополнение карт" -ForegroundColor White
Write-Host "      - Команды (/teams)" -ForegroundColor White
Write-Host "      - Фильтры транзакций" -ForegroundColor White
Write-Host "      - Реферальная программа (/referrals)" -ForegroundColor White
Write-Host ""
Write-Host "📊 Просмотр логов:" -ForegroundColor Cyan
Write-Host "   docker compose logs -f" -ForegroundColor White
Write-Host ""
