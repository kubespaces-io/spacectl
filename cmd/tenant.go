package cmd

import (
	"fmt"
	"os"
	"strings"

	"spacectl/internal/api"
	"spacectl/internal/models"

	"github.com/spf13/cobra"
)

// tenantCmd represents the tenant command
var tenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Manage tenants",
	Long:  `Manage Kubernetes tenants including listing, creating, updating, and deleting them.`,
}

func init() {
	rootCmd.AddCommand(tenantCmd)
}

// tenantListCmd represents the tenant list command
var tenantListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tenants",
	Long:  `List tenants. Use --project to filter by project.`,
	RunE:  runTenantList,
}

var tenantListProject string
var tenantListProjectName string
var tenantListAll bool

func init() {
	tenantCmd.AddCommand(tenantListCmd)
	tenantListCmd.Flags().StringVar(&tenantListProject, "project", "", "Project ID to filter tenants")
	tenantListCmd.Flags().StringVar(&tenantListProjectName, "project-name", "", "Project name to filter tenants")
	tenantListCmd.Flags().BoolVar(&tenantListAll, "all", false, "List tenants from all projects")
}

func runTenantList(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Validate flags
	if tenantListAll && (tenantListProject != "" || tenantListProjectName != "") {
		return fmt.Errorf("--all cannot be used with --project or --project-name")
	}
	if tenantListProjectName != "" && tenantListProject != "" {
		return fmt.Errorf("only one of --project or --project-name is allowed")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	tenantAPI := api.NewTenantAPI(client)

	if tenantListAll {
		// List tenants from all projects
		projectAPI := api.NewProjectAPI(client)
		userProjects, err := projectAPI.ListUserProjects()
		if err != nil {
			return fmt.Errorf("failed to list user projects: %w", err)
		}
		if len(userProjects) == 0 {
			return fmt.Errorf("no projects found. Create a project first")
		}

		// Create custom output for all tenants with proper alignment
		var output strings.Builder
		output.WriteString("PROJECT        NAME                CLOUD  REGION  VERSION   COMPUTE  MEMORY(GB)  STATUS\n")
		output.WriteString("-------        ----                -----  ------  -------   -------  ----------  ------\n")

		for _, membership := range userProjects {
			projectTenants, err := tenantAPI.ListProjectTenants(membership.Project.ID)
			if err != nil {
				return fmt.Errorf("failed to list tenants for project %s: %w", membership.Project.Name, err)
			}
			for _, tenant := range projectTenants {
				output.WriteString(fmt.Sprintf("%-15s %-20s %-6s %-7s %-9s %-8d %-11d %s\n",
					membership.Project.Name,
					tenant.Namespace,
					tenant.CloudProvider,
					tenant.Region,
					tenant.KubernetesVersion,
					tenant.ComputeQuota,
					tenant.MemoryQuotaGB,
					tenant.Status,
				))
			}
		}

		fmt.Print(output.String())
		return nil
	}

	// Single project logic
	if tenantListProject == "" && tenantListProjectName != "" {
		pid, err := resolveProjectID(client, tenantListProjectName, "", "")
		if err != nil {
			return err
		}
		tenantListProject = pid
	}

	// If still empty, use default project
	if tenantListProject == "" {
		// Get user's projects and use the first one as default
		projectAPI := api.NewProjectAPI(client)
		userProjects, err := projectAPI.ListUserProjects()
		if err != nil {
			return fmt.Errorf("failed to list user projects: %w", err)
		}
		if len(userProjects) == 0 {
			return fmt.Errorf("no projects found. Create a project first")
		}
		tenantListProject = userProjects[0].Project.ID
	}

	// Get tenants
	tenants, err := tenantAPI.ListProjectTenants(tenantListProject)
	if err != nil {
		return fmt.Errorf("failed to list tenants: %w", err)
	}

	// Output tenants
	return formatter.FormatData(tenants)
}

// tenantCreateCmd represents the tenant create command
var tenantCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a tenant",
	Long:  `Create a new Kubernetes tenant in the specified project.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTenantCreate,
}

var (
	tenantCreateProject         string
	tenantCreateProjectName     string
	tenantCreateCloud           string
	tenantCreateRegion          string
	tenantCreateK8sVersion      string
	tenantCreateCompute         int
	tenantCreateMemory          int
	tenantCreateNamespaceSuffix string
)

func init() {
	tenantCmd.AddCommand(tenantCreateCmd)
	tenantCreateCmd.Flags().StringVar(&tenantCreateProject, "project", "", "Project ID")
	tenantCreateCmd.Flags().StringVar(&tenantCreateProjectName, "project-name", "", "Project name")
	tenantCreateCmd.Flags().StringVar(&tenantCreateCloud, "cloud", "", "Cloud provider")
	tenantCreateCmd.Flags().StringVar(&tenantCreateRegion, "region", "", "Region")
	tenantCreateCmd.Flags().StringVar(&tenantCreateK8sVersion, "k8s-version", "", "Kubernetes version")
	tenantCreateCmd.Flags().IntVar(&tenantCreateCompute, "compute", 0, "Compute quota (cores)")
	tenantCreateCmd.Flags().IntVar(&tenantCreateMemory, "memory", 0, "Memory quota (GB)")
	tenantCreateCmd.Flags().StringVar(&tenantCreateNamespaceSuffix, "namespace-suffix", "", "Namespace suffix")
	tenantCreateCmd.MarkFlagRequired("project")
	tenantCreateCmd.MarkFlagRequired("cloud")
	tenantCreateCmd.MarkFlagRequired("region")
	tenantCreateCmd.MarkFlagRequired("k8s-version")
	tenantCreateCmd.MarkFlagRequired("compute")
	tenantCreateCmd.MarkFlagRequired("memory")
}

func runTenantCreate(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	name := args[0]

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	tenantAPI := api.NewTenantAPI(client)
	// Resolve project if name provided
	if tenantCreateProjectName != "" && tenantCreateProject != "" {
		return fmt.Errorf("only one of --project or --project-name is allowed")
	}
	if tenantCreateProject == "" && tenantCreateProjectName != "" {
		pid, err := resolveProjectID(client, tenantCreateProjectName, "", "")
		if err != nil {
			return err
		}
		tenantCreateProject = pid
	}

	// Prepare request
	req := models.CreateTenantRequest{
		Name:              name,
		CloudProvider:     tenantCreateCloud,
		Region:            tenantCreateRegion,
		KubernetesVersion: tenantCreateK8sVersion,
		ComputeQuota:      tenantCreateCompute,
		MemoryQuotaGB:     tenantCreateMemory,
		NamespaceSuffix:   tenantCreateNamespaceSuffix,
	}

	// Create tenant
	tenant, err := tenantAPI.CreateTenant(tenantCreateProject, req)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	// Output tenant
	return formatter.FormatData(tenant)
}

// tenantGetCmd represents the tenant get command
var tenantGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get tenant details",
	Long:  `Get detailed information about a specific tenant.`,
	Args:  cobra.NoArgs,
	RunE:  runTenantGet,
}

func init() {
	tenantCmd.AddCommand(tenantGetCmd)
}

var (
	tenantGetID          string
	tenantGetName        string
	tenantGetProjectID   string
	tenantGetProjectName string
)

func init() {
	tenantGetCmd.Flags().StringVar(&tenantGetID, "id", "", "Tenant ID")
	tenantGetCmd.Flags().StringVar(&tenantGetName, "name", "", "Tenant name")
	tenantGetCmd.Flags().StringVar(&tenantGetProjectID, "project", "", "Project ID (required if using --name)")
	tenantGetCmd.Flags().StringVar(&tenantGetProjectName, "project-name", "", "Project name (alternative to --project when using --name)")
}

func runTenantGet(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	tenantAPI := api.NewTenantAPI(client)
	// Resolve tenant
	if tenantGetName != "" && tenantGetID != "" {
		return fmt.Errorf("only one of --name or --id is allowed")
	}
	if tenantGetName != "" {
		// need project context
		if tenantGetProjectID != "" && tenantGetProjectName != "" {
			return fmt.Errorf("only one of --project or --project-name is allowed")
		}
		if tenantGetProjectID == "" && tenantGetProjectName != "" {
			pid, err := resolveProjectID(client, tenantGetProjectName, "", "")
			if err != nil {
				return err
			}
			tenantGetProjectID = pid
		}
		var err error
		tenantGetID, err = resolveTenantID(client, tenantGetName, "", tenantGetProjectID)
		if err != nil {
			return err
		}
	} else if tenantGetID == "" {
		return fmt.Errorf("either --name or --id must be provided")
	}

	// Get tenant
	tenant, err := tenantAPI.GetTenant(tenantGetID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Output tenant
	return formatter.FormatData(tenant)
}

// tenantDeleteCmd represents the tenant delete command
var tenantDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a tenant",
	Long:  `Delete a tenant. This action cannot be undone.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTenantDelete,
}

func init() {
	tenantCmd.AddCommand(tenantDeleteCmd)
}

func runTenantDelete(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	id := args[0]

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	tenantAPI := api.NewTenantAPI(client)

	// Delete tenant
	err := tenantAPI.DeleteTenant(id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Output success message
	if !quiet {
		fmt.Printf("Successfully deleted tenant %s\n", id)
	}

	return nil
}

// tenantStatusCmd represents the tenant status command
var tenantStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get tenant status",
	Long:  `Get the provisioning status of a tenant.`,
	Args:  cobra.NoArgs,
	RunE:  runTenantStatus,
}

var (
	tenantStatusID          string
	tenantStatusName        string
	tenantStatusProjectID   string
	tenantStatusProjectName string
)

func init() {
	tenantCmd.AddCommand(tenantStatusCmd)
	tenantStatusCmd.Flags().StringVar(&tenantStatusID, "id", "", "Tenant ID")
	tenantStatusCmd.Flags().StringVar(&tenantStatusName, "name", "", "Tenant name")
	tenantStatusCmd.Flags().StringVar(&tenantStatusProjectID, "project", "", "Project ID")
	tenantStatusCmd.Flags().StringVar(&tenantStatusProjectName, "project-name", "", "Project name")
}

func runTenantStatus(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	tenantAPI := api.NewTenantAPI(client)

	// Resolve tenant
	if tenantStatusName != "" && tenantStatusID != "" {
		return fmt.Errorf("only one of --name or --id is allowed")
	}
	if tenantStatusName != "" {
		// need project context
		if tenantStatusProjectID != "" && tenantStatusProjectName != "" {
			return fmt.Errorf("only one of --project or --project-name is allowed")
		}
		if tenantStatusProjectID == "" && tenantStatusProjectName != "" {
			pid, err := resolveProjectID(client, tenantStatusProjectName, "", "")
			if err != nil {
				return err
			}
			tenantStatusProjectID = pid
		}
		var err error
		tenantStatusID, err = resolveTenantID(client, tenantStatusName, "", tenantStatusProjectID)
		if err != nil {
			return err
		}
	} else if tenantStatusID == "" {
		return fmt.Errorf("either --name or --id must be provided")
	}

	// Get tenant status
	status, err := tenantAPI.GetTenantStatus(tenantStatusID)
	if err != nil {
		return fmt.Errorf("failed to get tenant status: %w", err)
	}

	// Output status
	return formatter.FormatData(status)
}

// tenantKubeconfigCmd represents the tenant kubeconfig command
var tenantKubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig <id>",
	Short: "Download tenant kubeconfig",
	Long:  `Download the kubeconfig file for a tenant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTenantKubeconfig,
}

var tenantKubeconfigOutputFile string

func init() {
	tenantCmd.AddCommand(tenantKubeconfigCmd)
	tenantKubeconfigCmd.Flags().StringVar(&tenantKubeconfigOutputFile, "output-file", "", "Output file path (default: stdout)")
}

func runTenantKubeconfig(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	id := args[0]

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	tenantAPI := api.NewTenantAPI(client)

	// Get kubeconfig
	kubeconfig, err := tenantAPI.GetTenantKubeconfig(id)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	// Output kubeconfig
	if tenantKubeconfigOutputFile != "" {
		err := os.WriteFile(tenantKubeconfigOutputFile, []byte(kubeconfig), 0600)
		if err != nil {
			return fmt.Errorf("failed to write kubeconfig file: %w", err)
		}
		if !quiet {
			fmt.Printf("Kubeconfig saved to %s\n", tenantKubeconfigOutputFile)
		}
	} else {
		fmt.Print(kubeconfig)
	}

	return nil
}

// tenantLocationsCmd represents the tenant locations command
var tenantLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "List available locations",
	Long:  `List available cloud provider and region combinations.`,
	RunE:  runTenantLocations,
}

func init() {
	tenantCmd.AddCommand(tenantLocationsCmd)
}

func runTenantLocations(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	tenantAPI := api.NewTenantAPI(client)

	// Get locations
	locations, err := tenantAPI.GetAvailableLocations()
	if err != nil {
		return fmt.Errorf("failed to get locations: %w", err)
	}

	// Output locations
	return formatter.FormatData(locations)
}

// tenantK8sVersionsCmd represents the tenant k8s-versions command
var tenantK8sVersionsCmd = &cobra.Command{
	Use:   "k8s-versions",
	Short: "List available Kubernetes versions",
	Long:  `List available Kubernetes versions for tenant creation.`,
	RunE:  runTenantK8sVersions,
}

func init() {
	tenantCmd.AddCommand(tenantK8sVersionsCmd)
}

func runTenantK8sVersions(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	tenantAPI := api.NewTenantAPI(client)

	// Get Kubernetes versions
	versions, err := tenantAPI.GetAvailableKubernetesVersions()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes versions: %w", err)
	}

	// Output versions
	return formatter.FormatData(versions)
}
