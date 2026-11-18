package model

import (
	"fmt"

	"local-sentiment-api/internal/sentiment"
)

func largestLogit(logits []float32) (int, error) {
	if len(logits) != 5 {
		return 0, fmt.Errorf("model returned %d logits; expected 5", len(logits))
	}
	best := 0
	for i := 1; i < len(logits); i++ {
		if logits[i] > logits[best] {
			best = i
		}
	}
	return best, nil
}

func starsToSentiment(starIndex int) sentiment.Label {
	switch starIndex {
	case 0, 1:
		return sentiment.Negative
	case 2:
		return sentiment.Neutral
	default:
		return sentiment.Positive
	}
}
