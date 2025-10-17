package cmd

import (
	"fmt"

	"spacectl/internal/version"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of spacectl.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("spacectl", version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
