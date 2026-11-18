package tokenizer

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEncodeUsesWordPiecesAndSpecialTokens(t *testing.T) {
	tokenizer := testTokenizer(t, "[PAD]\n[UNK]\n[CLS]\n[SEP]\nun\n##aff\n##able\n!\n")
	got, err := tokenizer.Encode("Unaffable!", 8)
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{2, 4, 5, 6, 7, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestEncodeTruncatesAndKeepsSeparator(t *testing.T) {
	tokenizer := testTokenizer(t, "[PAD]\n[UNK]\n[CLS]\n[SEP]\na\nb\nc\n")
	got, err := tokenizer.Encode("a b c", 4)
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{2, 4, 5, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestEncodeLowercasesAndStripsAccents(t *testing.T) {
	tokenizer := testTokenizer(t, "[PAD]\n[UNK]\n[CLS]\n[SEP]\ncafe\n")
	got, err := tokenizer.Encode("CAFÉ", 8)
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{2, 4, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func testTokenizer(t *testing.T, vocabulary string) *WordPiece {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vocab.txt")
	if err := os.WriteFile(path, []byte(vocabulary), 0o600); err != nil {
		t.Fatal(err)
	}
	tokenizer, err := LoadWordPieceVocabulary(path)
	if err != nil {
		t.Fatal(err)
	}
	return tokenizer
}
