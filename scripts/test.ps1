#
# Apache License 2.0
# Copyright (c) 2026 OTMC Softwares.
# Contributors: Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
#

Write-Host '╔══════════════════════════════════════════════════╗' -ForegroundColor Cyan
Write-Host '║              Test Manager v1.0                   ║' -ForegroundColor Cyan
Write-Host '╚══════════════════════════════════════════════════╝' -ForegroundColor Cyan

if ($args.Count -gt 0) {
    $option = $args[0]
} else {
    Write-Host "  1. Print logs" -ForegroundColor Green
    Write-Host "  2. Rotate logs" -ForegroundColor Green
    Write-Host "  3. Go test ./..." -ForegroundColor Green
    $option = Read-Host ">> Select option (1-3)"
}

switch ($option) {
    "1" {
        Set-Location $PSScriptRoot/..
        go run tests/printer/main.go
    }
    "2" {
        Set-Location $PSScriptRoot/..
        go run tests/rotator/main.go
    }
    "3" {
        Set-Location $PSScriptRoot/..
        go test -v ./...
    }
    default {
        Write-Host "Invalid option" -ForegroundColor Red
    }
}
