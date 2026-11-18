package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"local-sentiment-api/internal/sentiment"
)

type classifier map[string]sentiment.Label

func (c classifier) Classify(text string) (sentiment.Label, error) {
	label, ok := c[text]
	if !ok {
		return sentiment.Neutral, errors.New("missing test label")
	}
	return label, nil
}

func TestAnalyzeReturnsOnlyAggregate(t *testing.T) {
	handler := NewHandler(classifier{
		"good": sentiment.Positive,
		"bad":  sentiment.Negative,
		"okay": sentiment.Neutral,
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader(`{"sentences":["good","bad","okay"]}`))
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["total_analyzed"] != float64(3) || body["overall_score"] != float64(50) || body["overall_sentiment"] != "mixed" {
		t.Fatalf("unexpected response: %#v", body)
	}
	if _, found := body["results"]; found {
		t.Fatalf("response must not contain per-sentence results: %#v", body)
	}
}

func TestAnalyzeRejectsBlankSentence(t *testing.T) {
	handler := NewHandler(classifier{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader(`{"sentences":["  "]}`))
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestAnalyzeRejectsMoreThanFiveThousandSentences(t *testing.T) {
	sentences := make([]string, maxSentences+1)
	for i := range sentences {
		sentences[i] = "okay"
	}
	payload, err := json.Marshal(map[string]any{"sentences": sentences})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader(string(payload)))
	NewHandler(classifier{"okay": sentiment.Neutral}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "5000") {
		t.Fatalf("expected 5000-item validation error, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestHealth(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewHandler(classifier{}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected health response: %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestTesterPage(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewHandler(classifier{}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "CEA Sentiment Local") {
		t.Fatalf("unexpected tester page: %d %q", recorder.Code, recorder.Body.String())
	}
}
