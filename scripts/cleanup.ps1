# Генеральная уборка проекта
# Удаляет build артефакты и временные файлы

Write-Host "=== ГЕНЕРАЛЬНАЯ УБОРКА ПРОЕКТА ===" -ForegroundColor Cyan
Write-Host ""

$rootPath = "C:\Users\aalab\epn-killer-project"
Set-Location $rootPath

# Список папок для удаления
$foldersToDelete = @(
    "epn-killer-frontend\node_modules",
    "epn-killer-frontend\.expo",
    "epn-killer-frontend\final_build",
    "epn-killer-frontend\dist",
    "epn-killer-frontend\.git",
    "epn-killer-core\.expo",
    "epn-killer-core\dist",
    "epn-killer-core\.git"
)

foreach ($folder in $foldersToDelete) {
    $fullPath = Join-Path $rootPath $folder
    if (Test-Path $fullPath) {
        Write-Host "Удаляю: $folder" -ForegroundColor Yellow
        Remove-Item -Path $fullPath -Recurse -Force -ErrorAction SilentlyContinue
        Write-Host "  ✓ Удалено" -ForegroundColor Green
    } else {
        Write-Host "Пропускаю: $folder (не найдено)" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "=== УБОРКА ЗАВЕРШЕНА ===" -ForegroundColor Green
Write-Host ""
Write-Host "Сохранены:" -ForegroundColor Cyan
Write-Host "  - Исходный код" -ForegroundColor White
Write-Host "  - .env файлы" -ForegroundColor White
Write-Host "  - Конфигурационные файлы" -ForegroundColor White
Write-Host ""
