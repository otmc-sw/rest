#
# Apache License 2.0
# Copyright (c) 2026 OTMC Softwares.
# Contributors: Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
#

param(
    [string]$Action,
    [string]$TagName,
    [string]$Commit = 'HEAD'
)

function Write-Info($msg)     { Write-Host $msg -ForegroundColor Cyan }
function Write-Success($msg)  { Write-Host $msg -ForegroundColor Green }
function Write-Error($msg)    { Write-Host $msg -ForegroundColor Red }

function Get-CurrentBranch {
    return git rev-parse --abbrev-ref HEAD
}

function Confirm-DestructiveAction {
    param([string]$Message)
    Write-Host ''
    Write-Host $Message -ForegroundColor Yellow
    Write-Host ''

    $reply = Read-Host '❓ Do you want to continue? (Y/N)'
    if ($reply -ne 'Y') {
        Write-Host '✋ Operation cancelled.' -ForegroundColor Red
        exit 1
    }
}

function Show-Usage {
    Write-Host ''
    Write-Host '╔══════════════════════════════════════════════════╗' -ForegroundColor Cyan
    Write-Host '║              Git Tag Manager v1.0                ║' -ForegroundColor Cyan
    Write-Host '╚══════════════════════════════════════════════════╝' -ForegroundColor Cyan
    Write-Host ''
    Write-Host '  Usage:  tag.ps1 <action> <tag> [commit]' -ForegroundColor White
    Write-Host ''
    Write-Host '  ┌─ Actions ─────────────────────────────────────┐' -ForegroundColor Yellow
    Write-Host '  │' -ForegroundColor Yellow
    Write-Host '  │  b    Create / update a tag (force)' -ForegroundColor Green
    Write-Host '  │       Then force-push it to origin.' -ForegroundColor Gray
    Write-Host '  │' -ForegroundColor Yellow
    Write-Host '  │  r    Restore branch to a tag (force)' -ForegroundColor Green
    Write-Host '  │       Resets current branch to the tag and' -ForegroundColor Gray
    Write-Host '  │       force-pushes. Only works on main/master.' -ForegroundColor Gray
    Write-Host '  │' -ForegroundColor Yellow
    Write-Host '  └───────────────────────────────────────────────┘' -ForegroundColor Yellow
    Write-Host ''
    Write-Host '  ┌─ Examples ────────────────────────────────────┐' -ForegroundColor Yellow
    Write-Host '  │' -ForegroundColor Yellow
    Write-Host '  │  tag.ps1 b v0.1.5' -ForegroundColor Green
    Write-Host '  │  tag.ps1 b v0.1.5 abc1234' -ForegroundColor Green
    Write-Host '  │  tag.ps1 r v0.1.5' -ForegroundColor Green
    Write-Host '  │' -ForegroundColor Yellow
    Write-Host '  └───────────────────────────────────────────────┘' -ForegroundColor Yellow
    Write-Host ''
}

function Invoke-TagCreate {
    Write-Info "💡 Force creating tag '$TagName' at commit '$Commit'..."
    git tag -f $TagName $Commit

    Write-Info "⬆️  Force pushing tag '$TagName'..."
    git push origin $TagName --force

    Write-Success "✅ Tag '$TagName' created/updated!"
}

function Invoke-TagRestore {
    $branch = Get-CurrentBranch

    if ($branch -ne 'main' -and $branch -ne 'master') {
        Write-Error "Can only restore on 'main' or 'master'; current branch is '$branch'"
        exit 1
    }

    $commitCount = git rev-list --count "$TagName..$branch"

    if ($commitCount -gt 0) {
        Confirm-DestructiveAction @"
⚠️  WARNING: You are about to revert $commitCount commit(s) on branch '$branch'!
  Tag:     $TagName
  Branch:  $branch

This will run: git reset --hard $TagName AND force-push.
"@
    } else {
        Write-Info "No commits to revert. Branch is already at tag '$TagName'."
    }

    Write-Info "🔄 Resetting branch '$branch' to tag '$TagName'..."
    git reset --hard $TagName

    Write-Info "⬆️  Force pushing branch '$branch'..."
    git push origin $branch --force

    Write-Success "✅ Branch '$branch' reset to '$TagName'!"
}


Set-Location $PSScriptRoot/..
if (-not $Action -or $Action -eq 'h' -or $Action -eq '-h' -or $Action -eq '--help' -or $Action -eq 'help') {
    Show-Usage
    exit 0
}

switch ($Action) {
    'b' { Invoke-TagCreate }
    'r' { Invoke-TagRestore }
    default {
        Write-Error "Unknown action '$Action'"
        Show-Usage
        exit 1
    }
}

