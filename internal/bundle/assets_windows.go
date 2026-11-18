//go:build bundled && windows

package bundle

import _ "embed"

//go:embed payload/windows-amd64/sentiment.onnx
var modelAsset []byte

//go:embed payload/windows-amd64/vocab.txt
var vocabularyAsset []byte

//go:embed payload/windows-amd64/onnxruntime.dll
var runtimeAsset []byte

//go:embed payload/windows-amd64/onnxruntime_providers_shared.dll
var providersAsset []byte

const runtimeAssetName = "runtime/onnxruntime.dll"

var embeddedAssets = []asset{
	{name: "model/sentiment.onnx", data: modelAsset},
	{name: "model/vocab.txt", data: vocabularyAsset},
	{name: runtimeAssetName, data: runtimeAsset},
	{name: "runtime/onnxruntime_providers_shared.dll", data: providersAsset},
}
