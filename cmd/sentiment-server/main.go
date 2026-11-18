//go:build cgo && !bundled

package main

import (
	"flag"
	"log"
)

func main() {
	address := flag.String("addr", "127.0.0.1:18080", "HTTP listen address")
	modelPath := flag.String("model", "models/sentiment.onnx", "path to model_quantized.onnx")
	vocabularyPath := flag.String("vocab", "models/vocab.txt", "path to the model vocabulary")
	runtimePath := flag.String("runtime", "", "path to onnxruntime.dll or onnxruntime.so")
	maxTokens := flag.Int("max-tokens", 128, "maximum tokens per sentence (2-512)")
	flag.Parse()

	if err := serve(serverConfig{
		Address:        *address,
		ModelPath:      *modelPath,
		VocabularyPath: *vocabularyPath,
		RuntimePath:    *runtimePath,
		MaxTokens:      *maxTokens,
	}); err != nil {
		log.Fatal(err)
	}
}
