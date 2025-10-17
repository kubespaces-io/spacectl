package api

import (
	"fmt"
	"net/http"
	"spacectl/internal/models"
	"time"
)

// AuthAPI handles authentication-related API calls
type AuthAPI struct {
	client *Client
}

// NewAuthAPI creates a new AuthAPI
func NewAuthAPI(client *Client) *AuthAPI {
	return &AuthAPI{client: client}
}

// Login authenticates a user with email and password
func (a *AuthAPI) Login(email, password string) (*models.LoginResponse, error) {
	req := models.LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := a.client.doRequest("POST", "/api/v1/user/login", req)
	if err != nil {
		return nil, err
	}

	var loginResp models.LoginResponse
	if err := a.client.handleResponse(resp, &loginResp); err != nil {
		return nil, err
	}

	return &loginResp, nil
}

// Register registers a new user
func (a *AuthAPI) Register(email, password string) error {
	req := models.LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := a.client.doRequest("POST", "/api/v1/user/register", req)
	if err != nil {
		return err
	}

	return a.client.handleResponse(resp, nil)
}

// VerifyEmail verifies a user's email with a code
func (a *AuthAPI) VerifyEmail(email, code string) error {
	req := models.VerifyEmailRequest{
		Email: email,
		Code:  code,
	}

	resp, err := a.client.doRequest("POST", "/api/v1/user/verify", req)
	if err != nil {
		return err
	}

	return a.client.handleResponse(resp, nil)
}

// ResendVerificationCode resends a verification code
func (a *AuthAPI) ResendVerificationCode(email string) error {
	req := models.ResendVerificationRequest{
		Email: email,
	}

	resp, err := a.client.doRequest("POST", "/api/v1/user/verify/resend", req)
	if err != nil {
		return err
	}

	return a.client.handleResponse(resp, nil)
}

// GetUserInfo gets the current user's information
func (a *AuthAPI) GetUserInfo() (*models.User, error) {
	resp, err := a.client.doRequest("GET", "/api/v1/user/info", nil)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := a.client.handleResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdatePreferences updates user preferences
func (a *AuthAPI) UpdatePreferences(prefs *models.UserPreferences) error {
	resp, err := a.client.doRequest("PUT", "/api/v1/user/preferences", prefs)
	if err != nil {
		return err
	}

	return a.client.handleResponse(resp, nil)
}

// GetGithubAuthURL gets the GitHub OAuth authorization URL
func (a *AuthAPI) GetGithubAuthURL(callbackPort string) (string, error) {
	// Use a simple GET request to trigger the OAuth flow
	// The backend will redirect to GitHub with proper state handling
	url := "/api/v1/auth/github?cli=true"
	if callbackPort != "" {
		url += "&callback_port=" + callbackPort
	}

	// Create a custom HTTP client that doesn't follow redirects
	// so we can capture the Location header
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", a.client.baseURL+url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if a.client.config.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.client.config.AccessToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// The response should contain the GitHub OAuth URL in Location header
	if resp.StatusCode == 307 || resp.StatusCode == 302 {
		location := resp.Header.Get("Location")
		if location != "" {
			return location, nil
		}
	}

	return "", fmt.Errorf("failed to get GitHub OAuth URL: status %d", resp.StatusCode)
}

// HandleGithubCallback handles the GitHub OAuth callback
func (a *AuthAPI) HandleGithubCallback(code, state string) (*models.LoginResponse, error) {
	url := fmt.Sprintf("/api/v1/auth/github/callback?code=%s&state=%s", code, state)

	resp, err := a.client.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var loginResp models.LoginResponse
	if err := a.client.handleResponse(resp, &loginResp); err != nil {
		return nil, err
	}

	return &loginResp, nil
}
