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

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new Kubespaces account",
	Long: `Register a new Kubespaces account using your email and password.
If email and password are not provided as flags, you will be prompted for them.`,
	RunE: runRegister,
}

var (
	registerEmail    string
	registerPassword string
)

func init() {
	rootCmd.AddCommand(registerCmd)

	registerCmd.Flags().StringVar(&registerEmail, "email", "", "Email address")
	registerCmd.Flags().StringVar(&registerPassword, "password", "", "Password")
}

func runRegister(cmd *cobra.Command, args []string) error {
	// Get email if not provided
	if registerEmail == "" {
		fmt.Print("Email: ")
		reader := bufio.NewReader(os.Stdin)
		email, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read email: %w", err)
		}
		registerEmail = strings.TrimSpace(email)
	}

	// Get password if not provided
	if registerPassword == "" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // New line after password input
		registerPassword = string(passwordBytes)
	}

	// Create API client
    client := api.NewClient(cfg.APIURL, cfg, debug)
	authAPI := api.NewAuthAPI(client)

	// Attempt registration
	err := authAPI.Register(registerEmail, registerPassword)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	// Output success message
	if !quiet {
		fmt.Printf("Successfully registered %s. Please check your email for verification instructions.\n", registerEmail)
	}

	return nil
}
