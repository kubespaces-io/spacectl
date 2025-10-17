package cmd

import (
	"fmt"

	"spacectl/internal/api"

	"github.com/spf13/cobra"
)

// orgCmd represents the organization command
var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Manage organizations",
	Long:  `Manage organizations including listing, creating, updating, and deleting them.`,
}

func init() {
	rootCmd.AddCommand(orgCmd)
}

// orgListCmd represents the org list command
var orgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations",
	Long:  `List all organizations the current user belongs to.`,
	RunE:  runOrgList,
}

func init() {
	orgCmd.AddCommand(orgListCmd)
}

func runOrgList(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	orgAPI := api.NewOrganizationAPI(client)

	// Get organizations
	orgs, err := orgAPI.ListUserOrganizations()
	if err != nil {
		return fmt.Errorf("failed to list organizations: %w", err)
	}

	// Output organizations
	return formatter.FormatData(orgs)
}

// orgCreateCmd represents the org create command
var orgCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create an organization",
	Long:  `Create a new organization with the specified name.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runOrgCreate,
}

func init() {
	orgCmd.AddCommand(orgCreateCmd)
}

func runOrgCreate(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	name := args[0]

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	orgAPI := api.NewOrganizationAPI(client)

	// Create organization
	org, err := orgAPI.CreateOrganization(name)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	// Output organization
	return formatter.FormatData(org)
}

// orgGetCmd represents the org get command
var orgGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get organization details",
	Long:  `Get detailed information about a specific organization.`,
	Args:  cobra.NoArgs,
	RunE:  runOrgGet,
}

func init() {
	orgCmd.AddCommand(orgGetCmd)
}

var (
	orgGetName string
	orgGetID   string
)

func init() {
	orgGetCmd.Flags().StringVar(&orgGetName, "name", "", "Organization name")
	orgGetCmd.Flags().StringVar(&orgGetID, "id", "", "Organization ID")
}

func runOrgGet(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	orgAPI := api.NewOrganizationAPI(client)

	// Resolve organization
	resolvedID, err := resolveOrganizationID(client, orgGetName, orgGetID)
	if err != nil {
		return err
	}

	// Get organization
	org, err := orgAPI.GetOrganization(resolvedID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Output organization
	return formatter.FormatData(org)
}

// orgUpdateCmd represents the org update command
var orgUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an organization",
	Long:  `Update an organization's name.`,
	Args:  cobra.NoArgs,
	RunE:  runOrgUpdate,
}

var orgUpdateName string

func init() {
	orgCmd.AddCommand(orgUpdateCmd)
	orgUpdateCmd.Flags().StringVar(&orgUpdateName, "name", "", "New organization name")
	orgUpdateCmd.MarkFlagRequired("name")
}

var (
	orgUpdateTargetName string
	orgUpdateTargetID   string
)

func init() {
	orgUpdateCmd.Flags().StringVar(&orgUpdateTargetName, "org-name", "", "Organization name to update")
	orgUpdateCmd.Flags().StringVar(&orgUpdateTargetID, "org-id", "", "Organization ID to update")
}

func runOrgUpdate(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	orgAPI := api.NewOrganizationAPI(client)

	// Resolve organization to update
	resolvedID, err := resolveOrganizationID(client, orgUpdateTargetName, orgUpdateTargetID)
	if err != nil {
		return err
	}

	// Update organization
	org, err := orgAPI.UpdateOrganization(resolvedID, orgUpdateName)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	// Output organization
	return formatter.FormatData(org)
}

// orgDeleteCmd represents the org delete command
var orgDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an organization",
	Long:  `Delete an organization. This action cannot be undone.`,
	Args:  cobra.NoArgs,
	RunE:  runOrgDelete,
}

func init() {
	orgCmd.AddCommand(orgDeleteCmd)
}

var (
	orgDeleteName string
	orgDeleteID   string
)

func init() {
	orgDeleteCmd.Flags().StringVar(&orgDeleteName, "name", "", "Organization name")
	orgDeleteCmd.Flags().StringVar(&orgDeleteID, "id", "", "Organization ID")
}

func runOrgDelete(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	orgAPI := api.NewOrganizationAPI(client)

	// Resolve organization
	resolvedID, err := resolveOrganizationID(client, orgDeleteName, orgDeleteID)
	if err != nil {
		return err
	}

	// Delete organization
	err = orgAPI.DeleteOrganization(resolvedID)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	// Output success message
	if !quiet {
		fmt.Printf("Successfully deleted organization %s\n", resolvedID)
	}

	return nil
}

// orgSetDefaultCmd represents the org set-default command
var orgSetDefaultCmd = &cobra.Command{
	Use:   "set-default",
	Short: "Set default organization",
	Long:  `Set an organization as the user's default organization.`,
	Args:  cobra.NoArgs,
	RunE:  runOrgSetDefault,
}

func init() {
	orgCmd.AddCommand(orgSetDefaultCmd)
}

var (
	orgDefaultName string
	orgDefaultID   string
)

func init() {
	orgSetDefaultCmd.Flags().StringVar(&orgDefaultName, "name", "", "Organization name")
	orgSetDefaultCmd.Flags().StringVar(&orgDefaultID, "id", "", "Organization ID")
}

func runOrgSetDefault(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	orgAPI := api.NewOrganizationAPI(client)

	// Resolve organization
	resolvedID, err := resolveOrganizationID(client, orgDefaultName, orgDefaultID)
	if err != nil {
		return err
	}

	// Set default organization
	err = orgAPI.SetDefaultOrganization(resolvedID)
	if err != nil {
		return fmt.Errorf("failed to set default organization: %w", err)
	}

	// Output success message
	if !quiet {
		fmt.Printf("Successfully set organization %s as default\n", resolvedID)
	}

	return nil
}
