[CmdletBinding(SupportsShouldProcess = $true, ConfirmImpact = 'Medium')]
param(
    [Parameter(Mandatory = $true)]
    [ValidateNotNullOrEmpty()]
    [string]$TargetRoot,

    [switch]$Force,

    [switch]$IncludeReadme
)

# Use -WhatIf for a dry run (PowerShell's native ShouldProcess mechanism) —
# it prints what would be copied without touching the filesystem.

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$sourceRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path

git -C $sourceRoot rev-parse --is-inside-work-tree *>$null
if ($LASTEXITCODE -ne 0) {
    throw "Source root is not a git checkout: $sourceRoot`nThis script only copies git-tracked files, so it requires a git repository as its source."
}

$resolvedTarget = Resolve-Path -LiteralPath $TargetRoot -ErrorAction SilentlyContinue
if ($null -eq $resolvedTarget) {
    $targetRoot = (New-Item -ItemType Directory -Path $TargetRoot -Force).FullName
} else {
    $targetRoot = $resolvedTarget.Path
}

# Keep this list identical to the $items array in sync-boilerplate.sh.
$items = @(
    'BLUEPRINT.md'
    '.markdownlint.jsonc'
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

# Copies only git-tracked files under $RelativePath, so build artifacts and
# other .gitignore'd content sitting in a local checkout (e.g. node_modules/,
# dist/, __pycache__/) never leak into the sync target.
function Copy-TrackedDirectory {
    param(
        [string]$RelativePath
    )

    $destinationPath = Join-Path $targetRoot $RelativePath
    if (-not $Force -and (Test-Path -LiteralPath $destinationPath)) {
        throw "Destination already exists (use -Force to overwrite): $destinationPath"
    }

    $trackedFiles = git -C $sourceRoot ls-files -- $RelativePath
    foreach ($file in $trackedFiles) {
        if ([string]::IsNullOrWhiteSpace($file)) {
            continue
        }
        $sourceFile = Join-Path $sourceRoot $file
        $destinationFile = Join-Path $targetRoot $file
        if ($PSCmdlet.ShouldProcess($destinationFile, "Copy file from $sourceFile")) {
            $parentDirectory = Split-Path -Parent $destinationFile
            if ($parentDirectory -and -not (Test-Path -LiteralPath $parentDirectory)) {
                New-Item -ItemType Directory -Path $parentDirectory -Force | Out-Null
            }
            Copy-Item -LiteralPath $sourceFile -Destination $destinationFile -Force:$Force
        }
    }
}

foreach ($relativePath in $items) {
    $sourcePath = Join-Path $sourceRoot $relativePath
    if (-not (Test-Path -LiteralPath $sourcePath)) {
        throw "Missing source item: $relativePath"
    }

    # -Force is required here: on non-Windows platforms, PowerShell marks dotfiles/dotdirs
    # (.editorconfig, .github, etc.) as Hidden, and Get-Item excludes Hidden items by default
    # even though Test-Path above does not.
    $item = Get-Item -LiteralPath $sourcePath -Force
    if ($item.PSIsContainer) {
        Copy-TrackedDirectory -RelativePath $relativePath
        $copied.Add($relativePath) | Out-Null
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
