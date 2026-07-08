# OTMC License.
# Copyright (c) 2026 OTMC Softwares. All rights reserved.
# Contributors: Trung Ng, OTMC Authors.

Set-Location -Path $PSScriptRoot/..

$LicenseHeader = @'
/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
'@

$SrcDirs       = @(
    "./"
)
$IgnoredDirs   = @(
    "\logs\",
    "\docs\"
)
$Whitelist     = @(
    "Apache License",
    "Copyright",
    "Contributors:",
    "TODO: ",
    "go:embed"
)

function Should-KeepComment ([string]$Text) {
    foreach ($entry in $Whitelist) {
        if ($Text -match [regex]::Escape($entry)) { return $true }
    }
    return $false
}

function Is-IgnoredPath ([string]$FullPath) {
    foreach ($ignored in $IgnoredDirs) {
        if ($FullPath -like "*$ignored*") { return $true }
    }
    return $false
}

function Strip-BlockComments ([string]$Content) {
    return [regex]::Replace($Content, '^[ \t]*/\*.*?\*/[ \t]*(\r?\n|$)', {
        param($m) if (Should-KeepComment $m.Value) { $m.Value } else { "" }
    }, 'Singleline, Multiline')
}

function Strip-LineComments ([string]$Content) {
    $lines = $Content -split "`n"
    $result = @()

    foreach ($line in $lines) {
        # 1. Full-line // comment
        if ($line -match '^[ \t]*//') {
            if (Should-KeepComment $line) { $result += $line }
            continue
        }
        # 2. Skip strings (string safety)
        if ($line -match '["'']') {
            $result += $line
            continue
        }
        # 3. Inline // comment only after selective symbols , ) }
        if ($line -match '([,)}][^/]*?)\s*//') {
            $result += ($line -replace '\s*//.*$', '').TrimEnd()
            continue
        }
        $result += $line
    }
    return ($result -join "`n")
}

function Add-LicenseHeader ([string]$Content) {
    if ($Content -match '@License ') { return $Content }
    return "$LicenseHeader`n$Content"
}

function Process-GoFile ([string]$FilePath) {
    if (Is-IgnoredPath $FilePath) { return }

    $Content = Get-Content -Path $FilePath -Raw
    $Content = Add-LicenseHeader $Content
    $Content = Strip-BlockComments $Content
    $Content = Strip-LineComments $Content

    [System.IO.File]::WriteAllText($FilePath, $Content, [System.Text.Encoding]::UTF8)
}

foreach ($Dir in $SrcDirs) {
    if (-not (Test-Path $Dir)) {
        Write-Host "⚠️  Skipping missing directory: $Dir" -ForegroundColor Yellow
        continue
    }

    Write-Host "`n📁 Scanning Go files in: $Dir" -ForegroundColor Blue
    $Files = Get-ChildItem -Path $Dir -Recurse -Filter *.go | Where-Object { -not (Is-IgnoredPath $_.FullName) }

    foreach ($File in $Files) {
        $index = $Files.IndexOf($File) + 1
        Write-Host ("  → {0,3}/{1,3} 🌿 Processing: {2}" -f $index, $Files.Count, $File.Name)
        Process-GoFile -FilePath $File.FullName
    }
    Write-Host "🚀 Completed folder $Dir with $($Files.Count) files" -ForegroundColor Green
}

Write-Host "`n✨ Go comments clean up complete." -ForegroundColor Green

