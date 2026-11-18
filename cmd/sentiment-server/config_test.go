package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadOrCreateConfigCreatesDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), configFileName)
	config, err := readOrCreateConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.Port != defaultPort || config.Address() != "127.0.0.1:18080" {
		t.Fatalf("unexpected configuration: %+v", config)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "{\n  \"port\": 18080\n}\n" {
		t.Fatalf("unexpected starter configuration: %q", contents)
	}
}

func TestReadOrCreateConfigReadsExistingPort(t *testing.T) {
	path := filepath.Join(t.TempDir(), configFileName)
	if err := os.WriteFile(path, []byte("{\"port\": 19090}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := readOrCreateConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.Port != 19090 || config.Address() != "127.0.0.1:19090" {
		t.Fatalf("unexpected configuration: %+v", config)
	}
}

func TestReadOrCreateConfigRejectsInvalidPort(t *testing.T) {
	path := filepath.Join(t.TempDir(), configFileName)
	if err := os.WriteFile(path, []byte("{\"port\": 0}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readOrCreateConfig(path); err == nil {
		t.Fatal("expected invalid port error")
	}
}
