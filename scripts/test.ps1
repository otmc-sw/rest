#
# Apache License 2.0
# Copyright (c) 2026 OTMC Softwares.
# Contributors: Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
#

$TOP = $PSScriptRoot + "/.."

Write-Host '╔══════════════════════════════════════════════════╗' -ForegroundColor Cyan
Write-Host '║              Test Manager v1.0                   ║' -ForegroundColor Cyan
Write-Host '╚══════════════════════════════════════════════════╝' -ForegroundColor Cyan

if ($args.Count -gt 0) {
    $option = $args[0]
} else {
    Write-Host "  1. Run Fiber example" -ForegroundColor Green
    Write-Host "  2. Run Gin example" -ForegroundColor Green
    Write-Host "  3. Go test ./..." -ForegroundColor Green
    $option = Read-Host ">> Select option (1-3)"
}

switch ($option) {
    "1" {
        Set-Location $TOP/examples/fiber
        sqlc generate
        go mod tidy
        go build -o fiber-example.exe
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Go build failed" -ForegroundColor Red
            exit 1
        }
        & .\fiber-example.exe
    }
    "2" {
        Set-Location $TOP/examples/gin
        go run main.go
    }
    "3" {
        Set-Location $TOP
        go test -v ./...
    }
    default {
        Write-Host "Invalid option" -ForegroundColor Red
    }
}
