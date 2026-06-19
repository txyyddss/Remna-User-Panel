# Run from repository root: .\scripts\check-all.ps1
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Test-Path (Join-Path $root "frontend/node_modules"))) {
  Write-Host "npm install --prefix frontend" -ForegroundColor Cyan
  npm --prefix frontend install
}

Write-Host "go mod download" -ForegroundColor Cyan
go mod download

Write-Host "npm run check" -ForegroundColor Cyan
npm run check
