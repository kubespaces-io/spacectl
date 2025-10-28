package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadReturnsDefaultConfigWhenFileMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	defaultCfg := DefaultConfig()
	if !reflect.DeepEqual(cfg, defaultCfg) {
		t.Fatalf("expected default config %+v, got %+v", defaultCfg, cfg)
	}

	// Verify config file was not created as a side effect.
	configPath := filepath.Join(tmpDir, ".spacectl")
	if _, err := os.Stat(configPath); err == nil {
		t.Fatalf("expected config file %q not to exist", configPath)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &Config{
		APIURL:         "https://api.example.com",
		AccessToken:    "access-token",
		RefreshToken:   "refresh-token",
		UserEmail:      "user@example.com",
		DefaultCloud:   "gke",
		DefaultRegion:  "us-central1",
		DefaultCompute: 4,
		DefaultMemory:  16,
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if !reflect.DeepEqual(loaded, cfg) {
		t.Fatalf("loaded config %+v does not match saved config %+v", loaded, cfg)
	}
}

func TestAuthenticationHelpers(t *testing.T) {
	cfg := &Config{}

	if cfg.IsAuthenticated() {
		t.Fatalf("expected IsAuthenticated() to be false when tokens are empty")
	}

	cfg.UpdateTokens("new-access", "new-refresh", "user@example.com")
	if !cfg.IsAuthenticated() {
		t.Fatalf("expected IsAuthenticated() to be true after UpdateTokens")
	}
	if cfg.UserEmail != "user@example.com" {
		t.Fatalf("expected UserEmail to be updated, got %q", cfg.UserEmail)
	}

	cfg.ClearAuth()
	if cfg.IsAuthenticated() {
		t.Fatalf("expected IsAuthenticated() to be false after ClearAuth")
	}
	if cfg.UserEmail != "" {
		t.Fatalf("expected UserEmail to be cleared, got %q", cfg.UserEmail)
	}
}
