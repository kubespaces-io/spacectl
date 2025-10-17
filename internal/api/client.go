package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"spacectl/internal/config"
	"spacectl/internal/models"
)

// Client represents the API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	config     *config.Config
	debug      bool
}

// NewClient creates a new API client
func NewClient(baseURL string, cfg *config.Config, debug bool) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: cfg,
		debug:  debug,
	}
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	var debugBody []byte
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
		debugBody = jsonBody
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.config.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "[spacectl] -> %s %s\n", method, c.baseURL+path)
		if len(debugBody) > 0 {
			redacted := redactSensitiveJSON(debugBody)
			fmt.Fprintf(os.Stderr, "[spacectl]    body: %s\n", string(redacted))
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Handle 401 - try to refresh token
	if resp.StatusCode == http.StatusUnauthorized && c.config.RefreshToken != "" {
		resp.Body.Close()

		// Try to refresh token
		if err := c.refreshToken(); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}

		// Retry request with new token
		req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("retry request failed: %w", err)
		}
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "[spacectl] <- %s %s : %d\n", method, c.baseURL+path, resp.StatusCode)
	}

	return resp, nil
}

// redactSensitiveJSON masks sensitive fields in a JSON payload.
// It makes a best-effort attempt to redact common secrets like passwords and tokens.
func redactSensitiveJSON(raw []byte) []byte {
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		// If not JSON, return as-is
		return raw
	}
	redactRecursive(&v)
	redacted, err := json.Marshal(v)
	if err != nil {
		return raw
	}
	return redacted
}

func redactRecursive(v *interface{}) {
	switch val := (*v).(type) {
	case map[string]interface{}:
		for k, vv := range val {
			if isSensitiveKey(k) {
				val[k] = "***REDACTED***"
				continue
			}
			tmp := interface{}(vv)
			redactRecursive(&tmp)
			val[k] = tmp
		}
	case []interface{}:
		for i := range val {
			tmp := interface{}(val[i])
			redactRecursive(&tmp)
			val[i] = tmp
		}
	default:
		// primitives: nothing to do
	}
}

func isSensitiveKey(key string) bool {
	switch strings.ToLower(key) {
	case "password", "pass", "pwd", "access_token", "refresh_token", "token", "authorization":
		return true
	default:
		return false
	}
}

// refreshToken refreshes the access token using the refresh token
func (c *Client) refreshToken() error {
	// Build request directly to avoid recursive auto-refresh
	payload := models.RefreshTokenRequest{RefreshToken: c.config.RefreshToken}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token request: %w", err)
	}

	url := c.baseURL + "/api/v1/user/refresh"
	if c.debug {
		fmt.Fprintf(os.Stderr, "[spacectl] -> POST %s\n", url)
		fmt.Fprintf(os.Stderr, "[spacectl]    body: %s\n", string(redactSensitiveJSON(body)))
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if c.debug {
		fmt.Fprintf(os.Stderr, "[spacectl] <- POST %s : %d\n", url, resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		// Invalidate local tokens to avoid repeated failures
		c.config.ClearAuth()
		_ = c.config.Save()
		return fmt.Errorf("session expired (HTTP %d). Please run 'spacectl login' to re-authenticate", resp.StatusCode)
	}

	var loginResp models.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode refresh response: %w", err)
	}

	// Update config with new tokens
	c.config.UpdateTokens(loginResp.AccessToken, loginResp.RefreshToken, loginResp.User.Email)

	// Save updated config
	if err := c.config.Save(); err != nil {
		return fmt.Errorf("failed to save updated config: %w", err)
	}

	return nil
}

// handleResponse handles the HTTP response and returns appropriate error
func (c *Client) handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if result != nil {
			if err := json.Unmarshal(body, result); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}
		return nil
	}

	// Try to parse error response
	var errorResp models.ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err == nil {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, errorResp.Error)
	}

	return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
}

// IsAuthenticated returns true if the client has valid authentication
func (c *Client) IsAuthenticated() bool {
	return c.config.IsAuthenticated()
}
