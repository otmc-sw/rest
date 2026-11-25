param(
    [string]$t = "app"
)

$CURRENT_DIR = Get-Location
# Run example
try {
    Set-Location examples/$t
    
    Write-Host "🧩 Building OTMC Application..." -ForegroundColor Cyan
    go build -o main.exe
    if ($LastExitCode -ne 0) {
        Write-Host "❌ Failed to build OTMC Application" -ForegroundColor Red
        exit 1
    }
    Write-Host "🚀 Running OTMC Application..." -ForegroundColor Green
    .\main.exe
} finally {
    Set-Location $CURRENT_DIR
}
