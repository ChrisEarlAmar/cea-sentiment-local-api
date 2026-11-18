//go:build bundled && linux

package bundle

import _ "embed"

//go:embed payload/linux-amd64/sentiment.onnx
var modelAsset []byte

//go:embed payload/linux-amd64/vocab.txt
var vocabularyAsset []byte

//go:embed payload/linux-amd64/libonnxruntime.so
var runtimeAsset []byte

//go:embed payload/linux-amd64/libonnxruntime_providers_shared.so
var providersAsset []byte

const runtimeAssetName = "runtime/libonnxruntime.so"

var embeddedAssets = []asset{
	{name: "model/sentiment.onnx", data: modelAsset},
	{name: "model/vocab.txt", data: vocabularyAsset},
	{name: runtimeAssetName, data: runtimeAsset},
	{name: "runtime/libonnxruntime_providers_shared.so", data: providersAsset},
}
