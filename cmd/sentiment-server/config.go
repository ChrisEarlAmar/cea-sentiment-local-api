package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	configFileName = "cea-sentiment-local-api.json"
	defaultPort    = 18080
)

type appConfig struct {
	Port int `json:"port"`
}

func (c appConfig) Address() string {
	return fmt.Sprintf("127.0.0.1:%d", c.Port)
}

func loadOrCreateConfig() (appConfig, string, error) {
	executable, err := os.Executable()
	if err != nil {
		return appConfig{}, "", fmt.Errorf("locate executable: %w", err)
	}
	path := filepath.Join(filepath.Dir(executable), configFileName)
	config, err := readOrCreateConfig(path)
	return config, path, err
}

func readOrCreateConfig(path string) (appConfig, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := writeDefaultConfig(path); err != nil {
			return appConfig{}, err
		}
		return appConfig{Port: defaultPort}, nil
	}
	if err != nil {
		return appConfig{}, fmt.Errorf("open configuration %q: %w", path, err)
	}
	defer file.Close()

	var config appConfig
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return appConfig{}, fmt.Errorf("read configuration %q: %w", path, err)
	}
	if err := requireConfigEOF(decoder); err != nil {
		return appConfig{}, fmt.Errorf("read configuration %q: %w", path, err)
	}
	if config.Port < 1 || config.Port > 65535 {
		return appConfig{}, fmt.Errorf("configuration %q has invalid port %d; use a value from 1 to 65535", path, config.Port)
	}
	return config, nil
}

func writeDefaultConfig(path string) error {
	contents, err := json.MarshalIndent(appConfig{Port: defaultPort}, "", "  ")
	if err != nil {
		return err
	}
	contents = append(contents, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create configuration %q: %w", path, err)
	}
	defer file.Close()
	if _, err := file.Write(contents); err != nil {
		return fmt.Errorf("write configuration %q: %w", path, err)
	}
	return nil
}

func requireConfigEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("configuration has an unexpected JSON value")
	}
	return err
}
