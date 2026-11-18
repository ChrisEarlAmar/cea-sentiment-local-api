package model

import (
	"testing"

	"local-sentiment-api/internal/sentiment"
)

func TestLargestLogit(t *testing.T) {
	got, err := largestLogit([]float32{-1, 0.2, 0.1, 0.5, 0.4})
	if err != nil || got != 3 {
		t.Fatalf("got %d, %v", got, err)
	}
}

func TestLargestLogitRejectsUnexpectedClassCount(t *testing.T) {
	if _, err := largestLogit([]float32{0, 1}); err == nil {
		t.Fatal("expected error")
	}
}

func TestStarsToSentiment(t *testing.T) {
	cases := []struct {
		stars int
		want  sentiment.Label
	}{
		{0, sentiment.Negative},
		{1, sentiment.Negative},
		{2, sentiment.Neutral},
		{3, sentiment.Positive},
		{4, sentiment.Positive},
	}
	for _, test := range cases {
		if got := starsToSentiment(test.stars); got != test.want {
			t.Errorf("starsToSentiment(%d) = %d, want %d", test.stars, got, test.want)
		}
	}
}
