[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("fast", "full", "ci-quick", "hosted-full", "release", "deploy")]
    [string]$Profile,
    [string]$Workspace,
    [string]$Base,
    [string]$Head,
    [string]$FilesFrom,
    [switch]$Staged,
    [switch]$SkipInfra,
    [switch]$SkipDb,
    [switch]$WriteLocalReport
)

$ErrorActionPreference = "Stop"
$bash = Get-Command bash -ErrorAction SilentlyContinue
if (-not $bash) {
    throw "bash is required. Use Git for Windows, WSL, or the repository dev container."
}

$verifyArgs = @((Join-Path $PSScriptRoot "verify.sh"), "--profile", $Profile)
if ($Workspace) {
    $verifyArgs += @("--workspace", $Workspace)
}
if ($Staged) {
    $verifyArgs += "--staged"
}
if ($Base) {
    $verifyArgs += @("--base", $Base)
}
if ($Head) {
    $verifyArgs += @("--head", $Head)
}
if ($FilesFrom) {
    $verifyArgs += @("--files-from", $FilesFrom)
}
if ($SkipInfra) {
    $verifyArgs += "--skip-infra"
}
if ($SkipDb) {
    $verifyArgs += "--skip-db"
}
if ($WriteLocalReport) {
    $verifyArgs += "--write-local-report"
}

& $bash.Source @verifyArgs
exit $LASTEXITCODE
