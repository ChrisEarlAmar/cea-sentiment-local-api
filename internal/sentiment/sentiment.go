package sentiment

import (
	"errors"
	"math"
)

var ErrNoSentences = errors.New("sentences must contain at least one item")

type Label uint8

const (
	Negative Label = iota
	Neutral
	Positive
)

type Classifier interface {
	Classify(text string) (Label, error)
}

type Distribution struct {
	Positive float64 `json:"positive"`
	Negative float64 `json:"negative"`
	Neutral  float64 `json:"neutral"`
}

type Result struct {
	TotalAnalyzed    int          `json:"total_analyzed"`
	Sentiment        Distribution `json:"sentiment"`
	OverallScore     float64      `json:"overall_score"`
	OverallSentiment string       `json:"overall_sentiment"`
}

func Summarize(sentences []string, classifier Classifier) (Result, error) {
	if len(sentences) == 0 {
		return Result{}, ErrNoSentences
	}

	var positive, negative, neutral int
	for _, sentence := range sentences {
		label, err := classifier.Classify(sentence)
		if err != nil {
			return Result{}, err
		}
		switch label {
		case Positive:
			positive++
		case Negative:
			negative++
		case Neutral:
			neutral++
		default:
			return Result{}, errors.New("classifier returned an unknown sentiment label")
		}
	}

	total := float64(len(sentences))
	delta := float64(positive-negative) / total
	return Result{
		TotalAnalyzed: len(sentences),
		Sentiment: Distribution{
			Positive: percentage(positive, total),
			Negative: percentage(negative, total),
			Neutral:  percentage(neutral, total),
		},
		OverallScore:     roundOneDecimal(50 + 50*delta),
		OverallSentiment: overallSentiment(delta),
	}, nil
}

func percentage(count int, total float64) float64 {
	return roundOneDecimal(float64(count) / total * 100)
}

func overallSentiment(delta float64) string {
	switch {
	case delta >= 0.2:
		return "positive"
	case delta <= -0.2:
		return "negative"
	default:
		return "mixed"
	}
}

func roundOneDecimal(n float64) float64 {
	return math.Round(n*10) / 10
}
