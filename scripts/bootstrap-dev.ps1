[CmdletBinding()]
param(
    [switch]$CheckOnly,
    [switch]$Strict
)

$ErrorActionPreference = "Stop"
$workspace = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

foreach ($required in @("git", "bash", "node", "npm")) {
    if (-not (Get-Command $required -ErrorAction SilentlyContinue)) {
        throw "Required developer tool is missing: $required"
    }
}

$missing = @()
foreach ($optional in @("go", "python", "mvn", "docker", "terraform", "pwsh", "shellcheck")) {
    if (Get-Command $optional -ErrorAction SilentlyContinue) {
        Write-Host "available: $optional"
    }
    else {
        Write-Host "unavailable: $optional (some full/scoped checks cannot run)"
        $missing += $optional
    }
}

if (-not $CheckOnly) {
    & git -C $workspace config core.hooksPath .githooks
    if ($LASTEXITCODE -ne 0) {
        throw "Could not configure core.hooksPath"
    }
    Write-Host "Configured core.hooksPath=.githooks"
}

if ($Strict -and $missing.Count -gt 0) {
    throw "Strict bootstrap failed; missing: $($missing -join ', ')"
}

Write-Host "Developer bootstrap complete. No credentials or global packages changed."
