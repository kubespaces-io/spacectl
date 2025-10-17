package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"spacectl/internal/api"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Kubespaces",
	Long: `Login to Kubespaces using your email and password.
If email and password are not provided as flags, you will be prompted for them.

For GitHub OAuth authentication, use: spacectl auth login --github`,
	RunE: runLogin,
}

var (
	loginEmail          string
	loginPassword       string
	loginGithub         bool
	loginCallbackPort   string
)

func init() {
	authCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVar(&loginEmail, "email", "", "Email address")
	loginCmd.Flags().StringVar(&loginPassword, "password", "", "Password")
	loginCmd.Flags().BoolVar(&loginGithub, "github", false, "Use GitHub OAuth authentication")
	loginCmd.Flags().StringVar(&loginCallbackPort, "callback-port", "8081", "Port for OAuth callback server (used with --github)")
}

func runLogin(cmd *cobra.Command, args []string) error {
	// If --github flag is set, use GitHub OAuth
	if loginGithub {
		// Set the callback port for GitHub login
		githubCallbackPort = loginCallbackPort
		return runGithubLogin(cmd, args)
	}

	// Email/password login flow
	// Get email if not provided
	if loginEmail == "" {
		fmt.Print("Email: ")
		reader := bufio.NewReader(os.Stdin)
		email, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read email: %w", err)
		}
		loginEmail = strings.TrimSpace(email)
	}

	// Get password if not provided
	if loginPassword == "" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // New line after password input
		loginPassword = string(passwordBytes)
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	authAPI := api.NewAuthAPI(client)

	// Attempt login
	loginResp, err := authAPI.Login(loginEmail, loginPassword)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Update config with tokens
	cfg.UpdateTokens(loginResp.AccessToken, loginResp.RefreshToken, loginResp.User.Email)

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Output success message
	if !quiet {
		fmt.Printf("Successfully logged in as %s\n", loginResp.User.Email)
	}

	return nil
}
