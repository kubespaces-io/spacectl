package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"spacectl/internal/api"

	"github.com/spf13/cobra"
)

// githubLoginCmd represents the github-login command
var githubLoginCmd = &cobra.Command{
	Use:        "github-login",
	Short:      "Login to Kubespaces using GitHub OAuth",
	Deprecated: "Use 'spacectl auth login --github' instead",
	Long: `Login to Kubespaces using GitHub OAuth authentication.
This will open your browser to authenticate with GitHub and then return you to the CLI.

DEPRECATED: Use 'spacectl auth login --github' instead.`,
	RunE: runGithubLogin,
}

var (
	githubCallbackPort string
)

func init() {
	authCmd.AddCommand(githubLoginCmd)

	githubLoginCmd.Flags().StringVar(&githubCallbackPort, "callback-port", "8081", "Port for OAuth callback server")
}

func runGithubLogin(cmd *cobra.Command, args []string) error {
	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	authAPI := api.NewAuthAPI(client)

	// Create a channel to receive the tokens
	tokenChan := make(chan struct {
		accessToken  string
		refreshToken string
		userEmail    string
		err          error
	}, 1)

	// Start callback server
	server := startCallbackServer(githubCallbackPort, tokenChan)

	// Get GitHub OAuth URL
	authURL, err := authAPI.GetGithubAuthURL(githubCallbackPort)
	if err != nil {
		return fmt.Errorf("failed to get GitHub auth URL: %w", err)
	}

	// Open browser
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Please open this URL in your browser:\n%s\n", authURL)
	} else {
		fmt.Println("Opening browser for GitHub authentication...")
	}

	// Wait for callback
	fmt.Println("Waiting for GitHub authentication...")

	select {
	case result := <-tokenChan:
		// Shutdown the callback server
		if server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			server.Shutdown(ctx)
		}

		if result.err != nil {
			return fmt.Errorf("GitHub login failed: %w", result.err)
		}

		// Update config with tokens
		cfg.UpdateTokens(result.accessToken, result.refreshToken, result.userEmail)

		// Save config
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		// Output success message
		if !quiet {
			fmt.Printf("Successfully logged in as %s via GitHub\n", result.userEmail)
		}

		return nil

	case <-time.After(5 * time.Minute):
		if server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			server.Shutdown(ctx)
		}
		return fmt.Errorf("authentication timeout - please try again")
	}
}

func startCallbackServer(port string, tokenChan chan<- struct {
	accessToken  string
	refreshToken string
	userEmail    string
	err          error
}) *http.Server {
	mux := http.NewServeMux()

	// Handle POST /callback - receives tokens from backend via JavaScript
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers to allow requests from any origin (since this is a local callback server)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse token response from backend
		var tokenResponse struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			UserEmail    string `json:"user_email"`
		}

		if err := json.NewDecoder(r.Body).Decode(&tokenResponse); err != nil {
			tokenChan <- struct {
				accessToken  string
				refreshToken string
				userEmail    string
				err          error
			}{err: fmt.Errorf("failed to parse token response: %w", err)}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Send tokens to main goroutine
		tokenChan <- struct {
			accessToken  string
			refreshToken string
			userEmail    string
			err          error
		}{
			accessToken:  tokenResponse.AccessToken,
			refreshToken: tokenResponse.RefreshToken,
			userEmail:    tokenResponse.UserEmail,
			err:          nil,
		}

		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	// Handle GET / - fallback/health check
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>spacectl GitHub Login</title>
</head>
<body>
    <h1>Waiting for GitHub authentication...</h1>
    <p>Please complete the authentication in the GitHub OAuth window.</p>
</body>
</html>`)
	})

	// Start server
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		tokenChan <- struct {
			accessToken  string
			refreshToken string
			userEmail    string
			err          error
		}{err: fmt.Errorf("failed to start callback server: %w", err)}
		return nil
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Serve in background
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			tokenChan <- struct {
				accessToken  string
				refreshToken string
				userEmail    string
				err          error
			}{err: fmt.Errorf("callback server error: %w", err)}
		}
	}()

	return server
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
