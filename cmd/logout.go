package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Kubespaces",
	Long:  `Logout from Kubespaces by clearing stored authentication tokens.`,
	RunE:  runLogout,
}

func init() {
	authCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	// Clear authentication tokens
	cfg.ClearAuth()

	// Save updated config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Output success message
	if !quiet {
		fmt.Println("Successfully logged out")
	}

	return nil
}
