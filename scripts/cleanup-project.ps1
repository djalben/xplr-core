# PowerShell Script for EPN Killer Project Cleanup
# Removes build artifacts and temporary files while preserving source code

param(
    [switch]$RemoveGit = $false,
    [switch]$Force = $false
)

Write-Host "=== EPN Killer Project Cleanup Script ===" -ForegroundColor Cyan
Write-Host ""

# Define root directory
$rootDir = Split-Path -Parent $PSScriptRoot

# Define folders to clean
$foldersToRemove = @(
    "node_modules",
    ".expo",
    "final_build",
    "dist",
    "build",
    ".next",
    "out",
    ".cache"
)

# Function to remove folders recursively
function Remove-Folders {
    param(
        [string]$basePath,
        [string[]]$folderNames
    )

    foreach ($folderName in $folderNames) {
        $folders = Get-ChildItem -Path $basePath -Filter $folderName -Recurse -Directory -ErrorAction SilentlyContinue

        foreach ($folder in $folders) {
            try {
                Write-Host "Removing: $($folder.FullName)" -ForegroundColor Yellow
                Remove-Item -Path $folder.FullName -Recurse -Force -ErrorAction Stop
                Write-Host "  ✓ Removed successfully" -ForegroundColor Green
            }
            catch {
                Write-Host "  ✗ Failed to remove: $_" -ForegroundColor Red
            }
        }
    }
}

# Confirm action
if (-not $Force) {
    Write-Host "This script will remove the following folders:" -ForegroundColor Yellow
    $foldersToRemove | ForEach-Object { Write-Host "  - $_" }

    if ($RemoveGit) {
        Write-Host "  - .git (Git repository)" -ForegroundColor Red
    }

    Write-Host ""
    $confirm = Read-Host "Continue? (y/n)"

    if ($confirm -ne "y") {
        Write-Host "Cleanup cancelled." -ForegroundColor Yellow
        exit
    }
}

Write-Host ""
Write-Host "Starting cleanup..." -ForegroundColor Cyan
Write-Host ""

# Remove build artifacts
Remove-Folders -basePath $rootDir -folderNames $foldersToRemove

# Remove .git if requested
if ($RemoveGit) {
    $gitPath = Join-Path $rootDir ".git"
    if (Test-Path $gitPath) {
        try {
            Write-Host "Removing Git repository: $gitPath" -ForegroundColor Yellow
            Remove-Item -Path $gitPath -Recurse -Force -ErrorAction Stop
            Write-Host "  ✓ Git repository removed" -ForegroundColor Green
        }
        catch {
            Write-Host "  ✗ Failed to remove Git repository: $_" -ForegroundColor Red
        }
    }
}

Write-Host ""
Write-Host "=== Cleanup Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Preserved files:" -ForegroundColor Cyan
Write-Host "  ✓ Source code (.ts, .tsx, .go, etc.)" -ForegroundColor Green
Write-Host "  ✓ Configuration files (.env, .json, .yaml)" -ForegroundColor Green
Write-Host "  ✓ Documentation (.md)" -ForegroundColor Green
