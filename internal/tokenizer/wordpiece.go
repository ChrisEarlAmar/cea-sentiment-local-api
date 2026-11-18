package tokenizer

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const maxWordPieceRunes = 100

type WordPiece struct {
	vocabulary map[string]int64
	unknown    int64
	cls        int64
	sep        int64
}

func LoadWordPieceVocabulary(path string) (*WordPiece, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	vocabulary := make(map[string]int64)
	scanner := bufio.NewScanner(file)
	for id := int64(0); scanner.Scan(); id++ {
		vocabulary[scanner.Text()] = id
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	unknown, okUnknown := vocabulary["[UNK]"]
	cls, okCLS := vocabulary["[CLS]"]
	sep, okSEP := vocabulary["[SEP]"]
	if !okUnknown || !okCLS || !okSEP {
		return nil, errors.New("vocabulary is missing required BERT special tokens")
	}
	return &WordPiece{vocabulary: vocabulary, unknown: unknown, cls: cls, sep: sep}, nil
}

func (w *WordPiece) Encode(text string, maxLength int) ([]int64, error) {
	if maxLength < 2 {
		return nil, fmt.Errorf("max length must be at least 2, got %d", maxLength)
	}

	ids := make([]int64, 0, maxLength)
	ids = append(ids, w.cls)
	for _, token := range basicTokens(text) {
		for _, id := range w.wordPieces(token) {
			if len(ids) == maxLength-1 {
				ids = append(ids, w.sep)
				return ids, nil
			}
			ids = append(ids, id)
		}
	}
	return append(ids, w.sep), nil
}

func (w *WordPiece) wordPieces(token string) []int64 {
	chars := []rune(token)
	if len(chars) > maxWordPieceRunes {
		return []int64{w.unknown}
	}

	pieces := make([]int64, 0, len(chars))
	for start := 0; start < len(chars); {
		end := len(chars)
		var id int64
		found := false
		for start < end {
			piece := string(chars[start:end])
			if start > 0 {
				piece = "##" + piece
			}
			if candidate, ok := w.vocabulary[piece]; ok {
				id, found = candidate, true
				break
			}
			end--
		}
		if !found {
			return []int64{w.unknown}
		}
		pieces = append(pieces, id)
		start = end
	}
	return pieces
}

func basicTokens(text string) []string {
	var cleaned strings.Builder
	for _, r := range text {
		switch {
		case r == 0 || r == 0xfffd || (unicode.IsControl(r) && !unicode.IsSpace(r)):
			continue
		case unicode.IsSpace(r):
			cleaned.WriteByte(' ')
		case isChinese(r):
			cleaned.WriteByte(' ')
			cleaned.WriteRune(r)
			cleaned.WriteByte(' ')
		default:
			cleaned.WriteRune(r)
		}
	}

	tokens := make([]string, 0)
	for _, word := range strings.Fields(cleaned.String()) {
		word = stripAccents(strings.ToLower(word))
		tokens = append(tokens, splitPunctuation(word)...)
	}
	return tokens
}

func stripAccents(text string) string {
	var stripped strings.Builder
	for _, r := range norm.NFD.String(text) {
		if !unicode.Is(unicode.Mn, r) {
			stripped.WriteRune(r)
		}
	}
	return stripped.String()
}

func splitPunctuation(text string) []string {
	tokens := make([]string, 0)
	var word strings.Builder
	flush := func() {
		if word.Len() > 0 {
			tokens = append(tokens, word.String())
			word.Reset()
		}
	}
	for _, r := range text {
		if isPunctuation(r) {
			flush()
			tokens = append(tokens, string(r))
			continue
		}
		word.WriteRune(r)
	}
	flush()
	return tokens
}

func isPunctuation(r rune) bool {
	return (r >= '!' && r <= '/') || (r >= ':' && r <= '@') || (r >= '[' && r <= '`') || (r >= '{' && r <= '~') || unicode.IsPunct(r)
}

func isChinese(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || (r >= 0x3400 && r <= 0x4DBF) || (r >= 0x20000 && r <= 0x2A6DF) || (r >= 0x2A700 && r <= 0x2B73F) || (r >= 0x2B740 && r <= 0x2B81F) || (r >= 0x2B820 && r <= 0x2CEAF) || (r >= 0xF900 && r <= 0xFAFF) || (r >= 0x2F800 && r <= 0x2FA1F)
}
