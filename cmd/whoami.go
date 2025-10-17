package cmd

import (
	"fmt"

	"spacectl/internal/api"

	"github.com/spf13/cobra"
)

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current user information",
	Long:  `Display information about the currently authenticated user.`,
	RunE:  runWhoami,
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func runWhoami(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
    client := api.NewClient(cfg.APIURL, cfg, debug)
	authAPI := api.NewAuthAPI(client)

	// Get user info
	user, err := authAPI.GetUserInfo()
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	// Output user info
	return formatter.FormatData(user)
}
