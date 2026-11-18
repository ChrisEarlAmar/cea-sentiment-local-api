package api

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"local-sentiment-api/internal/sentiment"
)

//go:embed web/index.html
var testerPage []byte

const (
	maxRequestBytes = 5 << 20
	maxSentences    = 5000
)

type analyzeRequest struct {
	Sentences []string `json:"sentences"`
}

func NewHandler(classifier sentiment.Classifier) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", tester)
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("POST /api/analyze", analyze(classifier))
	return mux
}

func tester(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(testerPage)
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func analyze(classifier sentiment.Classifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := http.MaxBytesReader(w, r.Body, maxRequestBytes)
		defer body.Close()

		var request analyzeRequest
		decoder := json.NewDecoder(body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "request must be a JSON object with a sentences array")
			return
		}
		if err := requireEOF(decoder); err != nil {
			writeError(w, http.StatusBadRequest, "request must contain one JSON object")
			return
		}
		if err := validateSentences(request.Sentences); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		result, err := sentiment.Summarize(request.Sentences, classifier)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "sentiment model could not analyze the request")
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("unexpected JSON value")
	}
	return err
}

func validateSentences(sentences []string) error {
	if len(sentences) == 0 {
		return sentiment.ErrNoSentences
	}
	if len(sentences) > maxSentences {
		return errors.New("sentences may contain at most 5000 items")
	}
	for _, sentence := range sentences {
		if strings.TrimSpace(sentence) == "" {
			return errors.New("sentences may not contain blank items")
		}
	}
	return nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
