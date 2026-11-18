#!/usr/bin/env sh
set -eu

project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
destination="$project_root/internal/bundle/payload/linux-amd64"

for source in models/sentiment.onnx models/vocab.txt runtime/libonnxruntime.so runtime/libonnxruntime_providers_shared.so; do
  if [ ! -f "$project_root/$source" ]; then
    echo "Missing $source. Run 'sh scripts/fetch-assets.sh' first." >&2
    exit 1
  fi
done

mkdir -p "$destination"
cp "$project_root/models/sentiment.onnx" "$destination/sentiment.onnx"
cp "$project_root/models/vocab.txt" "$destination/vocab.txt"
cp "$project_root/runtime/libonnxruntime.so" "$destination/libonnxruntime.so"
cp "$project_root/runtime/libonnxruntime_providers_shared.so" "$destination/libonnxruntime_providers_shared.so"

echo "Prepared embedded Linux x64 assets."
