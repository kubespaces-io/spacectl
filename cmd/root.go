package cmd

import (
	"fmt"
	"os"

	"spacectl/internal/config"
	"spacectl/internal/output"

	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	apiURL    string
	outputFmt string
	noHeaders bool
	quiet     bool
    debug     bool
	cfg       *config.Config
	formatter *output.Formatter
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "spacectl",
	Short: "A CLI tool for managing Kubespaces resources",
	Long: `spacectl is a command-line tool for managing Kubespaces resources including
organizations, projects, and tenants. It provides a simple interface to interact
with the Kubespaces API.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override API URL if provided
		if apiURL != "" {
			cfg.APIURL = apiURL
		}

        // Create formatter
		format := output.Format(outputFmt)
		formatter = output.NewFormatter(format, noHeaders, os.Stdout)

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.spacectl)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "API URL (overrides config)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "Output format (table, json, yaml, csv)")
	rootCmd.PersistentFlags().BoolVar(&noHeaders, "no-headers", false, "Suppress headers in table/CSV output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Minimal output")
    rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging of API requests")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		// Note: This is not implemented yet as we use a fixed config path
		fmt.Printf("Using config file: %s\n", cfgFile)
	}
}
