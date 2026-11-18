//go:build cgo

package model

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	ort "github.com/yalue/onnxruntime_go"

	"local-sentiment-api/internal/sentiment"
	"local-sentiment-api/internal/tokenizer"
)

const defaultMaxTokens = 128

type ONNXClassifier struct {
	mu        sync.Mutex
	tokenizer *tokenizer.WordPiece
	session   *ort.DynamicAdvancedSession
	maxTokens int
	closed    bool
}

func NewONNXClassifier(modelPath, vocabularyPath, runtimePath string, maxTokens int) (*ONNXClassifier, error) {
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	}
	if maxTokens < 2 || maxTokens > 512 {
		return nil, fmt.Errorf("max tokens must be between 2 and 512, got %d", maxTokens)
	}

	wordPiece, err := tokenizer.LoadWordPieceVocabulary(vocabularyPath)
	if err != nil {
		return nil, fmt.Errorf("load vocabulary: %w", err)
	}
	if runtimePath == "" {
		runtimePath = defaultRuntimePath()
	}

	ort.SetSharedLibraryPath(runtimePath)
	if err := ort.InitializeEnvironment(ort.WithLogLevelWarning()); err != nil {
		return nil, fmt.Errorf("initialize ONNX Runtime from %q: %w", runtimePath, err)
	}
	initialized := true
	defer func() {
		if initialized {
			_ = ort.DestroyEnvironment()
		}
	}()

	options, err := ort.NewSessionOptions()
	if err != nil {
		return nil, fmt.Errorf("create ONNX session options: %w", err)
	}
	defer options.Destroy()
	if err := options.SetIntraOpNumThreads(1); err != nil {
		return nil, fmt.Errorf("set ONNX CPU threads: %w", err)
	}

	session, err := ort.NewDynamicAdvancedSession(
		modelPath,
		[]string{"input_ids", "attention_mask", "token_type_ids"},
		[]string{"logits"},
		options,
	)
	if err != nil {
		return nil, fmt.Errorf("load ONNX model %q: %w", modelPath, err)
	}

	initialized = false
	return &ONNXClassifier{
		tokenizer: wordPiece,
		session:   session,
		maxTokens: maxTokens,
	}, nil
}

func (c *ONNXClassifier) Classify(text string) (sentiment.Label, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return sentiment.Neutral, fmt.Errorf("ONNX classifier is closed")
	}

	ids, err := c.tokenizer.Encode(text, c.maxTokens)
	if err != nil {
		return sentiment.Neutral, fmt.Errorf("tokenize text: %w", err)
	}
	shape := ort.NewShape(1, int64(len(ids)))
	attentionMask := make([]int64, len(ids))
	for i := range attentionMask {
		attentionMask[i] = 1
	}
	tokenTypeIDs := make([]int64, len(ids))

	inputIDs, err := ort.NewTensor(shape, ids)
	if err != nil {
		return sentiment.Neutral, fmt.Errorf("create input_ids tensor: %w", err)
	}
	defer inputIDs.Destroy()
	attention, err := ort.NewTensor(shape, attentionMask)
	if err != nil {
		return sentiment.Neutral, fmt.Errorf("create attention_mask tensor: %w", err)
	}
	defer attention.Destroy()
	tokenTypes, err := ort.NewTensor(shape, tokenTypeIDs)
	if err != nil {
		return sentiment.Neutral, fmt.Errorf("create token_type_ids tensor: %w", err)
	}
	defer tokenTypes.Destroy()

	inputs := []ort.Value{inputIDs, attention, tokenTypes}
	outputs := []ort.Value{nil}
	if err := c.session.Run(inputs, outputs); err != nil {
		return sentiment.Neutral, fmt.Errorf("run ONNX model: %w", err)
	}
	if outputs[0] == nil {
		return sentiment.Neutral, fmt.Errorf("model returned no logits")
	}
	defer outputs[0].Destroy()

	logits, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return sentiment.Neutral, fmt.Errorf("model logits have unexpected type %T", outputs[0])
	}
	starIndex, err := largestLogit(logits.GetData())
	if err != nil {
		return sentiment.Neutral, err
	}
	return starsToSentiment(starIndex), nil
}

func (c *ONNXClassifier) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	sessionErr := c.session.Destroy()
	environmentErr := ort.DestroyEnvironment()
	if sessionErr != nil {
		return sessionErr
	}
	return environmentErr
}

func defaultRuntimePath() string {
	name := "libonnxruntime.so"
	if runtime.GOOS == "windows" {
		name = "onnxruntime.dll"
	}
	return filepath.Join("runtime", name)
}
