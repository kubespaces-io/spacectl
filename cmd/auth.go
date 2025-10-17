package cmd

import (
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long: `Authentication commands for managing login and logout.
Supports both email/password and GitHub OAuth authentication.`,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
