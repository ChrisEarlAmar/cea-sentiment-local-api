[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

function Get-SHA256([string]$Path) {
    $algorithm = [System.Security.Cryptography.SHA256]::Create()
    $stream = [System.IO.File]::OpenRead($Path)
    try {
        return ([System.BitConverter]::ToString($algorithm.ComputeHash($stream))).Replace('-', '').ToLowerInvariant()
    } finally {
        $stream.Dispose()
        $algorithm.Dispose()
    }
}

$projectRoot = Split-Path -Parent $PSScriptRoot
$buildToolsRoot = Join-Path $projectRoot 'build-tools'
$toolRoot = Join-Path $buildToolsRoot 'w64devkit'
$gccPath = Join-Path $toolRoot 'bin\gcc.exe'

if (Test-Path -LiteralPath $gccPath) {
    Write-Output (Split-Path -Parent $gccPath)
    exit 0
}

$archiveName = 'w64devkit-x64-2.11.0.7z.exe'
$archivePath = Join-Path $buildToolsRoot $archiveName
$expectedSHA256 = '54c2bb9962e8a30c8270e34c5efc3d5fee579f53087fa7d435b95fbba2af7b2b'
$downloadUrl = "https://github.com/skeeto/w64devkit/releases/download/v2.11.0/$archiveName"

New-Item -ItemType Directory -Force -Path $buildToolsRoot | Out-Null
if (-not (Test-Path -LiteralPath $archivePath)) {
    Write-Host 'Downloading the one-time Windows build compiler…' -ForegroundColor Cyan
    Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath
}

$actualSHA256 = Get-SHA256 $archivePath
if ($actualSHA256 -ne $expectedSHA256) {
    throw "The downloaded build compiler failed its checksum check: $archivePath"
}

Write-Host 'Preparing the one-time Windows build compiler…' -ForegroundColor Cyan
$sevenZip = 'C:\Program Files\7-Zip\7z.exe'
if (Test-Path -LiteralPath $sevenZip) {
    & $sevenZip x '-y' ("-o{0}" -f $buildToolsRoot) $archivePath | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'Could not extract the build compiler.' }
} else {
    & $archivePath '-y' ("-o{0}" -f $buildToolsRoot)
    if ($LASTEXITCODE -ne 0) { throw 'Could not extract the build compiler.' }
}

if (-not (Test-Path -LiteralPath $gccPath)) {
    throw "The extracted build compiler did not contain GCC at $gccPath"
}

Write-Output (Split-Path -Parent $gccPath)
