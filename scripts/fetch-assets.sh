#!/usr/bin/env sh
set -eu

runtime_version="${1:-1.30.0}"
project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
models_dir="$project_root/models"
runtime_dir="$project_root/runtime"
model_path="$models_dir/sentiment.onnx"
vocabulary_path="$models_dir/vocab.txt"
model_sha256="b969ce8faf498c40096d0be82b0a1180ebc3ae470ed7eed940345bd646a5dce7"

if [ "$(uname -s)" != "Linux" ] || [ "$(uname -m)" != "x86_64" ]; then
  echo "This script supports Linux x86_64. Use fetch-assets.ps1 on Windows x64." >&2
  exit 1
fi

for command in curl unzip sha256sum; do
  command -v "$command" >/dev/null || { echo "Missing required command: $command" >&2; exit 1; }
done

mkdir -p "$models_dir" "$runtime_dir"
if [ ! -f "$model_path" ]; then
  curl --fail --location --retry 3 --output "$model_path" \
    "https://huggingface.co/onnx-community/bert-base-multilingual-uncased-sentiment-ONNX/resolve/main/onnx/model_quantized.onnx"
fi
if [ ! -f "$vocabulary_path" ]; then
  curl --fail --location --retry 3 --output "$vocabulary_path" \
    "https://huggingface.co/onnx-community/bert-base-multilingual-uncased-sentiment-ONNX/raw/main/vocab.txt"
fi

actual_hash=$(sha256sum "$model_path" | awk '{print $1}')
if [ "$actual_hash" != "$model_sha256" ]; then
  echo "Unexpected model hash for $model_path. Delete it and run this script again." >&2
  exit 1
fi

temporary_dir=$(mktemp -d)
trap 'rm -rf "$temporary_dir"' EXIT
package_path="$temporary_dir/onnxruntime.nupkg"
curl --fail --location --retry 3 --output "$package_path" \
  "https://api.nuget.org/v3-flatcontainer/microsoft.ml.onnxruntime/$runtime_version/microsoft.ml.onnxruntime.$runtime_version.nupkg"
unzip -q "$package_path" 'runtimes/linux-x64/native/*' -d "$temporary_dir/package"
install -m 755 "$temporary_dir/package/runtimes/linux-x64/native/libonnxruntime.so" "$runtime_dir/libonnxruntime.so"
install -m 755 "$temporary_dir/package/runtimes/linux-x64/native/libonnxruntime_providers_shared.so" "$runtime_dir/libonnxruntime_providers_shared.so"

echo "Downloaded the model and ONNX Runtime to models/ and runtime/."
