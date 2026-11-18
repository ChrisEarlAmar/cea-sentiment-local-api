package sentiment

import (
	"errors"
	"testing"
)

type stubClassifier map[string]Label

func (s stubClassifier) Classify(text string) (Label, error) {
	label, ok := s[text]
	if !ok {
		return Neutral, errors.New("missing test label")
	}
	return label, nil
}

func TestSummarize(t *testing.T) {
	classifier := stubClassifier{
		"good":     Positive,
		"bad":      Negative,
		"ordinary": Neutral,
	}

	got, err := Summarize([]string{"good", "bad", "ordinary", "good"}, classifier)
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalAnalyzed != 4 || got.Sentiment.Positive != 50 || got.Sentiment.Negative != 25 || got.Sentiment.Neutral != 25 {
		t.Fatalf("unexpected distribution: %+v", got)
	}
	if got.OverallScore != 62.5 || got.OverallSentiment != "positive" {
		t.Fatalf("unexpected overall result: %+v", got)
	}
}

func TestSummarizeMixed(t *testing.T) {
	got, err := Summarize([]string{"good", "bad"}, stubClassifier{
		"good": Positive,
		"bad":  Negative,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.OverallScore != 50 || got.OverallSentiment != "mixed" {
		t.Fatalf("unexpected mixed result: %+v", got)
	}
}

func TestSummarizeRejectsEmptyInput(t *testing.T) {
	_, err := Summarize(nil, stubClassifier{})
	if !errors.Is(err, ErrNoSentences) {
		t.Fatalf("expected ErrNoSentences, got %v", err)
	}
}
