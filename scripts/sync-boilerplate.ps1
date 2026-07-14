[CmdletBinding(SupportsShouldProcess = $true, ConfirmImpact = 'Medium')]
param(
    [Parameter(Mandatory = $true)]
    [ValidateNotNullOrEmpty()]
    [string]$TargetRoot,

    [switch]$Force,

    [switch]$IncludeReadme
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$sourceRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path

$resolvedTarget = Resolve-Path -LiteralPath $TargetRoot -ErrorAction SilentlyContinue
if ($null -eq $resolvedTarget) {
    $targetRoot = (New-Item -ItemType Directory -Path $TargetRoot -Force).FullName
} else {
    $targetRoot = $resolvedTarget.Path
}

# Keep this list identical to the $items array in sync-boilerplate.sh.
$items = @(
    'BLUEPRINT.md'
    'AGENTS.md'
    'CONTRIBUTING.md'
    'docs'
    'LICENSE'
    'SECURITY.md'
    '.editorconfig'
    '.gitignore'
    '.gitmessage'
    '.github'
    'templates'
    'scripts'
)

if ($IncludeReadme) {
    $items = @('README.md', 'README.ko.md', 'README.ja.md') + $items
}

$copied = New-Object System.Collections.Generic.List[string]

foreach ($relativePath in $items) {
    $sourcePath = Join-Path $sourceRoot $relativePath
    if (-not (Test-Path -LiteralPath $sourcePath)) {
        throw "Missing source item: $relativePath"
    }

    $item = Get-Item -LiteralPath $sourcePath
    if ($item.PSIsContainer) {
        $destinationPath = Join-Path $targetRoot $relativePath
        if (-not $Force -and (Test-Path -LiteralPath $destinationPath)) {
            throw "Destination already exists (use -Force to overwrite): $destinationPath"
        }
        if ($PSCmdlet.ShouldProcess($destinationPath, "Copy directory from $sourcePath")) {
            New-Item -ItemType Directory -Path $destinationPath -Force | Out-Null
            Copy-Item -LiteralPath (Join-Path $sourcePath '*') -Destination $destinationPath -Recurse -Force:$Force
            $copied.Add($relativePath) | Out-Null
        }
        continue
    }

    $destinationPath = Join-Path $targetRoot $relativePath
    if (-not $Force -and (Test-Path -LiteralPath $destinationPath)) {
        throw "Destination already exists (use -Force to overwrite): $destinationPath"
    }
    $parentDirectory = Split-Path -Parent $destinationPath
    if ($parentDirectory -and -not (Test-Path -LiteralPath $parentDirectory)) {
        New-Item -ItemType Directory -Path $parentDirectory -Force | Out-Null
    }

    if ($PSCmdlet.ShouldProcess($destinationPath, "Copy file from $sourcePath")) {
        Copy-Item -LiteralPath $sourcePath -Destination $destinationPath -Force:$Force
        $copied.Add($relativePath) | Out-Null
    }
}

Write-Host 'Convention sync completed.'
Write-Host "Target root: $targetRoot"
Write-Host 'Copied items:'
$copied | ForEach-Object { Write-Host " - $_" }
