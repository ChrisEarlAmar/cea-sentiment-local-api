//go:build cgo

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"local-sentiment-api/internal/api"
	"local-sentiment-api/internal/model"
)

type serverConfig struct {
	Address        string
	ModelPath      string
	VocabularyPath string
	RuntimePath    string
	MaxTokens      int
	OpenBrowser    bool
}

func serve(config serverConfig) error {
	classifier, err := model.NewONNXClassifier(config.ModelPath, config.VocabularyPath, config.RuntimePath, config.MaxTokens)
	if err != nil {
		return err
	}
	defer classifier.Close()

	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		return err
	}
	defer listener.Close()

	server := &http.Server{
		Handler:           api.NewHandler(classifier),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if config.OpenBrowser {
		go func() {
			if err := openBrowser("http://" + listener.Addr().String()); err != nil {
				log.Printf("open tester page: %v", err)
			}
		}()
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}
	}()

	log.Printf("sentiment API and tester page listening on http://%s", listener.Addr())
	if err := server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP API: %w", err)
	}
	return nil
}
