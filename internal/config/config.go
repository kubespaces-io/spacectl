package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the spacectl configuration
type Config struct {
	APIURL       string `json:"api_url"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserEmail    string `json:"user_email"`

	// Default tenant creation settings
	DefaultCloud   string `json:"default_cloud,omitempty"`
	DefaultRegion  string `json:"default_region,omitempty"`
	DefaultCompute int    `json:"default_compute,omitempty"`
	DefaultMemory  int    `json:"default_memory,omitempty"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		APIURL:         "http://localhost:8080",
		DefaultCloud:   "eks",
		DefaultRegion:  "eu",
		DefaultCompute: 2,
		DefaultMemory:  4,
	}
}

// Load loads the configuration from ~/.spacectl
func Load() (*Config, error) {
	configPath := getConfigPath()

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save saves the configuration to ~/.spacectl
func (c *Config) Save() error {
	configPath := getConfigPath()

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// IsAuthenticated returns true if the user has valid tokens
func (c *Config) IsAuthenticated() bool {
	return c.AccessToken != "" && c.RefreshToken != ""
}

// ClearAuth clears authentication tokens
func (c *Config) ClearAuth() {
	c.AccessToken = ""
	c.RefreshToken = ""
	c.UserEmail = ""
}

// UpdateTokens updates the access and refresh tokens
func (c *Config) UpdateTokens(accessToken, refreshToken, userEmail string) {
	c.AccessToken = accessToken
	c.RefreshToken = refreshToken
	c.UserEmail = userEmail
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory is not available
		return ".spacectl"
	}
	return filepath.Join(homeDir, ".spacectl")
}
