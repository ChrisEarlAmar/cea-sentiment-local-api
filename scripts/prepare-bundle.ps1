$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $PSScriptRoot
$source = @{
    "models\sentiment.onnx" = "sentiment.onnx"
    "models\vocab.txt" = "vocab.txt"
    "runtime\onnxruntime.dll" = "onnxruntime.dll"
    "runtime\onnxruntime_providers_shared.dll" = "onnxruntime_providers_shared.dll"
}
$destination = Join-Path $projectRoot "internal\bundle\payload\windows-amd64"
New-Item -ItemType Directory -Force -Path $destination | Out-Null

foreach ($entry in $source.GetEnumerator()) {
    $from = Join-Path $projectRoot $entry.Key
    if (-not (Test-Path -LiteralPath $from)) {
        throw "Missing $from. Run .\scripts\fetch-assets.ps1 first."
    }
    Copy-Item -LiteralPath $from -Destination (Join-Path $destination $entry.Value) -Force
}

Write-Host "Prepared embedded Windows x64 assets."
