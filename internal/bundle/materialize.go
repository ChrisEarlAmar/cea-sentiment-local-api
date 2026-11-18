//go:build bundled

package bundle

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Paths struct {
	ModelPath      string
	VocabularyPath string
	RuntimePath    string
}

type asset struct {
	name string
	data []byte
}

func Materialize() (Paths, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return Paths{}, fmt.Errorf("find user cache directory: %w", err)
	}
	root := filepath.Join(cacheDir, "local-sentiment-api", "bundled-v1")
	for _, embedded := range embeddedAssets {
		if err := writeAsset(root, embedded); err != nil {
			return Paths{}, err
		}
	}
	return Paths{
		ModelPath:      filepath.Join(root, "model", "sentiment.onnx"),
		VocabularyPath: filepath.Join(root, "model", "vocab.txt"),
		RuntimePath:    filepath.Join(root, runtimeAssetName),
	}, nil
}

func writeAsset(root string, embedded asset) error {
	destination := filepath.Join(root, embedded.name)
	expected := sha256.Sum256(embedded.data)
	if actual, err := hashFile(destination); err == nil && actual == expected {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return fmt.Errorf("create asset cache directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(destination), ".sentiment-*")
	if err != nil {
		return fmt.Errorf("create temporary asset: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(embedded.data); err != nil {
		temporary.Close()
		return fmt.Errorf("write %s: %w", embedded.name, err)
	}
	if err := temporary.Chmod(0o700); err != nil {
		temporary.Close()
		return fmt.Errorf("set permissions on %s: %w", embedded.name, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close %s: %w", embedded.name, err)
	}
	if err := os.Remove(destination); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("replace %s: %w", embedded.name, err)
	}
	if err := os.Rename(temporaryPath, destination); err != nil {
		return fmt.Errorf("install %s: %w", embedded.name, err)
	}
	return nil
}

func hashFile(path string) ([sha256.Size]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return [sha256.Size]byte{}, err
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}
