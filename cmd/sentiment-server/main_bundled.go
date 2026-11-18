//go:build cgo && bundled

package main

import (
	"flag"
	"log"

	"local-sentiment-api/internal/bundle"
)

func main() {
	maxTokens := flag.Int("max-tokens", 128, "maximum tokens per sentence (2-512)")
	noBrowser := flag.Bool("no-browser", false, "do not open the local tester page")
	flag.Parse()

	configuration, configPath, err := loadOrCreateConfig()
	if err != nil {
		startupFailure(err)
	}
	log.Printf("using configuration from %s", configPath)

	assets, err := bundle.Materialize()
	if err != nil {
		startupFailure(err)
	}
	if err := serve(serverConfig{
		Address:        configuration.Address(),
		ModelPath:      assets.ModelPath,
		VocabularyPath: assets.VocabularyPath,
		RuntimePath:    assets.RuntimePath,
		MaxTokens:      *maxTokens,
		OpenBrowser:    !*noBrowser,
	}); err != nil {
		startupFailure(err)
	}
}
