# dev.ps1
if (Get-Command air -ErrorAction SilentlyContinue) {
    Write-Host "🚀 Starting Live Reload (Air)..." -ForegroundColor Magenta
    air
} else {
    Write-Host "⚠️  Air not found. Running with go run..." -ForegroundColor Yellow
    Write-Host "💡 To get live-reloading, run: go install github.com/air-verse/air@latest" -ForegroundColor Cyan
    go run ./cmd/api/main.go
}
