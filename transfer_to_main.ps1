# ============================================================
# Перенос рабочего проекта в C:\Users\aalab\epn-killer-project
# Запуск: из корня воркспейса (ifs): .\transfer_to_main.ps1
# ============================================================

$Source = $PSScriptRoot
$Dest   = "C:\Users\aalab\epn-killer-project"

if (-not (Test-Path $Source)) {
    Write-Error "Source not found: $Source"
    exit 1
}

Write-Host "Copying project from:" $Source
Write-Host "To:" $Dest
Write-Host ""

# Exclude .git and other optional dirs to avoid overwriting main repo
$Exclude = @(".git", "node_modules", "__pycache__", ".cursor")
$Items = Get-ChildItem -Path $Source -Force | Where-Object { $Exclude -notcontains $_.Name }

New-Item -ItemType Directory -Force -Path $Dest | Out-Null

foreach ($Item in $Items) {
    $Target = Join-Path $Dest $Item.Name
    if ($Item.PSIsContainer) {
        Write-Host "Copying folder:" $Item.Name
        Copy-Item -Path $Item.FullName -Destination $Target -Recurse -Force
    } else {
        Copy-Item -Path $Item.FullName -Destination $Target -Force
    }
}

Write-Host ""
Write-Host "Done. Next: cd $Dest ; copy .env.example .env ; docker compose up -d"
