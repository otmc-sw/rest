#
# Apache License 2.0
# Copyright (c) 2026 OTMC Softwares.
# Contributors: Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
#

$TOP = $PSScriptRoot + "/.."

Write-Host '╔══════════════════════════════════════════════════╗' -ForegroundColor Cyan
Write-Host '║              Test Manager v1.0                   ║' -ForegroundColor Cyan
Write-Host '╚══════════════════════════════════════════════════╝' -ForegroundColor Cyan

function Show-Menu {
    Write-Host "  1. Run Fiber example" -ForegroundColor Green
    Write-Host "  2. Test Fiber example" -ForegroundColor Green
    Write-Host "  3. Run Playwright tests" -ForegroundColor Green
    Write-Host "  4. Go test ./..." -ForegroundColor Green
}

if ($args.Count -gt 0) {
    $option = $args[0]
} else {
    Show-Menu
    $option = Read-Host ">> Select option (1-4)"
}

switch ($option) {
    "1" {
        Set-Location $TOP/examples/fiber
        sqlc generate
        go mod tidy
        go build -o fiber.exe
        if ($LASTEXITCODE -ne 0) {
            Write-Host "ERROR: Go build failed with exit code $LASTEXITCODE" -ForegroundColor Red
            exit 1
        }
        & .\fiber.exe
    }
    "2" {
        Set-Location $TOP/examples/fiber
        go test -v ./...
    }
    "3" {
        Set-Location $TOP/tests/playwright
        if (-not (Test-Path "node_modules")) {
            npm install
        }
        npx playwright test
    }
    "4" {
        Set-Location $TOP
        go test -v ./...
    }
    default {
        Write-Host "ERROR: Invalid option provided: '$option'" -ForegroundColor Red
        exit 1
    }
}
