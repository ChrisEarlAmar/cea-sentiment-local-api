[CmdletBinding()]
param(
    [string]$RuntimeVersion = "1.30.0"
)

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $PSScriptRoot
$modelsDir = Join-Path $projectRoot "models"
$runtimeDir = Join-Path $projectRoot "runtime"
$modelPath = Join-Path $modelsDir "sentiment.onnx"
$vocabularyPath = Join-Path $modelsDir "vocab.txt"
$modelSHA256 = "B969CE8FAF498C40096D0BE82B0A1180EBC3AE470ED7EED940345BD646A5DCE7"

New-Item -ItemType Directory -Force -Path $modelsDir, $runtimeDir | Out-Null

function Get-Download([string]$Url, [string]$Destination) {
    if (-not (Test-Path -LiteralPath $Destination)) {
        Invoke-WebRequest -Uri $Url -OutFile $Destination
    }
}

function Get-SHA256([string]$Path) {
    $algorithm = [System.Security.Cryptography.SHA256]::Create()
    $stream = [System.IO.File]::OpenRead($Path)
    try {
        return ([System.BitConverter]::ToString($algorithm.ComputeHash($stream))).Replace("-", "")
    }
    finally {
        $stream.Dispose()
        $algorithm.Dispose()
    }
}

Get-Download "https://huggingface.co/onnx-community/bert-base-multilingual-uncased-sentiment-ONNX/resolve/main/onnx/model_quantized.onnx" $modelPath
Get-Download "https://huggingface.co/onnx-community/bert-base-multilingual-uncased-sentiment-ONNX/raw/main/vocab.txt" $vocabularyPath

$actualHash = Get-SHA256 $modelPath
if ($actualHash -ne $modelSHA256) {
    throw "Unexpected model hash for $modelPath. Delete it and run this script again."
}

$packagePath = Join-Path $env:TEMP "Microsoft.ML.OnnxRuntime.$RuntimeVersion.nupkg"
$expandedPath = Join-Path $env:TEMP "Microsoft.ML.OnnxRuntime.$RuntimeVersion"
Get-Download "https://api.nuget.org/v3-flatcontainer/microsoft.ml.onnxruntime/$RuntimeVersion/microsoft.ml.onnxruntime.$RuntimeVersion.nupkg" $packagePath
Expand-Archive -LiteralPath $packagePath -DestinationPath $expandedPath -Force

$nativePath = Join-Path $expandedPath "runtimes\win-x64\native"
foreach ($file in @("onnxruntime.dll", "onnxruntime_providers_shared.dll")) {
    Copy-Item -LiteralPath (Join-Path $nativePath $file) -Destination (Join-Path $runtimeDir $file) -Force
}

Write-Host "Downloaded the model and ONNX Runtime to models\\ and runtime\\."
