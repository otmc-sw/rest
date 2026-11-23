
$CURRENT_DIR = Get-Location
# Run example
try {
    Set-Location examples/otmc
    Write-Host "🧩 Building OTMC Application..." -ForegroundColor Cyan
    go build -o main.exe
    Write-Host "🚀 Running OTMC Application..." -ForegroundColor Green
    .\main.exe
} finally {
    Set-Location $CURRENT_DIR
}
