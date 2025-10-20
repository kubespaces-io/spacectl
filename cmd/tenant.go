package cmd

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	Use:   "delete",
	Short: "Delete a tenant",
	Long:  `Delete a tenant. This action cannot be undone.`,
	Args:  cobra.NoArgs,
	RunE:  runTenantDelete,
}

var (
	tenantDeleteForce       bool
	tenantDeleteID          string
	tenantDeleteName        string
	tenantDeleteProjectID   string
	tenantDeleteProjectName string
)

func init() {
	tenantCmd.AddCommand(tenantDeleteCmd)
	tenantDeleteCmd.Flags().BoolVar(&tenantDeleteForce, "force", false, "Skip confirmation prompt")
	tenantDeleteCmd.Flags().StringVar(&tenantDeleteID, "id", "", "Tenant ID")
	tenantDeleteCmd.Flags().StringVar(&tenantDeleteName, "name", "", "Tenant name")
	tenantDeleteCmd.Flags().StringVar(&tenantDeleteProjectID, "project", "", "Project ID (required if using --name)")
	tenantDeleteCmd.Flags().StringVar(&tenantDeleteProjectName, "project-name", "", "Project name (alternative to --project when using --name)")
}

func runTenantDelete(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	tenantAPI := api.NewTenantAPI(client)

	// Resolve tenant
	if tenantDeleteName != "" && tenantDeleteID != "" {
		return fmt.Errorf("only one of --name or --id is allowed")
	}
	if tenantDeleteName != "" {
		// need project context
		if tenantDeleteProjectID != "" && tenantDeleteProjectName != "" {
			return fmt.Errorf("only one of --project or --project-name is allowed")
		}
		if tenantDeleteProjectID == "" && tenantDeleteProjectName != "" {
			pid, err := resolveProjectID(client, tenantDeleteProjectName, "", "")
			if err != nil {
				return err
			}
			tenantDeleteProjectID = pid
		}
		var err error
		tenantDeleteID, err = resolveTenantID(client, tenantDeleteName, "", tenantDeleteProjectID)
		if err != nil {
			return err
		}
	} else if tenantDeleteID == "" {
		return fmt.Errorf("either --name or --id must be provided")
	}

	// Get tenant details for confirmation
	tenant, err := tenantAPI.GetTenant(tenantDeleteID)
	if err != nil {
		return fmt.Errorf("failed to get tenant details: %w", err)
	}

	// Ask for confirmation unless --force is used
	if !tenantDeleteForce {
		fmt.Printf("Are you sure you want to delete tenant '%s' (ID: %s)? This action cannot be undone.\n", tenant.Name, tenantDeleteID)
		fmt.Print("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Delete tenant
	err = tenantAPI.DeleteTenant(tenantDeleteID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Output success message
	if !quiet {
		fmt.Printf("Successfully deleted tenant %s\n", tenantDeleteID)
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

// tenantKubectlCmd represents the tenant kubectl command
var tenantKubectlCmd = &cobra.Command{
	Use:   "kubectl [flags] -- [kubectl args]",
	Short: "Execute kubectl commands on a tenant",
	Long: `Execute kubectl commands on a tenant using its kubeconfig.
The kubeconfig is automatically retrieved and cached for performance.

Examples:
  spacectl tenant kubectl --name my-tenant --project my-project -- get pods
  spacectl tenant kubectl --id abc123 -- get nodes
  spacectl tenant kubectl --name my-tenant --project my-project -- apply -f deployment.yaml`,
	RunE:                   runTenantKubectl,
	DisableFlagsInUseLine:  true,
	DisableFlagParsing:     false,
	FParseErrWhitelist:     cobra.FParseErrWhitelist{UnknownFlags: true},
}

var (
	tenantKubectlName      string
	tenantKubectlID        string
	tenantKubectlProjectID string
	tenantKubectlProjectName string
	tenantKubectlNoCache   bool
)

func init() {
	tenantCmd.AddCommand(tenantKubectlCmd)
	tenantKubectlCmd.Flags().StringVar(&tenantKubectlName, "name", "", "Tenant name")
	tenantKubectlCmd.Flags().StringVar(&tenantKubectlID, "id", "", "Tenant ID")
	tenantKubectlCmd.Flags().StringVar(&tenantKubectlProjectID, "project", "", "Project ID (required if using --name)")
	tenantKubectlCmd.Flags().StringVar(&tenantKubectlProjectName, "project-name", "", "Project name (alternative to --project)")
	tenantKubectlCmd.Flags().BoolVar(&tenantKubectlNoCache, "no-cache", false, "Skip cache and fetch fresh kubeconfig")
}

func runTenantKubectl(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Parse arguments to find the separator "--"
	var kubectlArgs []string
	var spacectlArgs []string
	foundSeparator := false

	for i, arg := range args {
		if arg == "--" {
			foundSeparator = true
			spacectlArgs = args[:i]
			if i+1 < len(args) {
				kubectlArgs = args[i+1:]
			}
			break
		}
	}

	if !foundSeparator {
		kubectlArgs = args
	}

	if len(kubectlArgs) == 0 {
		return fmt.Errorf("no kubectl command provided. Usage: spacectl tenant kubectl [flags] -- <kubectl-command>")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	tenantAPI := api.NewTenantAPI(client)

	// Resolve tenant ID
	var tenantID string
	var err error

	if tenantKubectlName != "" && tenantKubectlID != "" {
		return fmt.Errorf("only one of --name or --id is allowed")
	}

	if tenantKubectlName != "" {
		// Need project context for name resolution
		if tenantKubectlProjectID != "" && tenantKubectlProjectName != "" {
			return fmt.Errorf("only one of --project or --project-name is allowed")
		}
		if tenantKubectlProjectID == "" && tenantKubectlProjectName != "" {
			pid, err := resolveProjectID(client, tenantKubectlProjectName, "", "")
			if err != nil {
				return err
			}
			tenantKubectlProjectID = pid
		}
		if tenantKubectlProjectID == "" {
			return fmt.Errorf("--project or --project-name is required when using --name")
		}

		tenantID, err = resolveTenantID(client, tenantKubectlName, "", tenantKubectlProjectID)
		if err != nil {
			return err
		}
	} else if tenantKubectlID != "" {
		tenantID = tenantKubectlID
	} else {
		return fmt.Errorf("either --name or --id must be provided")
	}

	// Get or retrieve kubeconfig
	kubeconfigPath, err := getOrFetchKubeconfig(tenantAPI, tenantID, tenantKubectlNoCache)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	// Execute kubectl with the kubeconfig
	kubectlCmd := exec.Command("kubectl", kubectlArgs...)
	kubectlCmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath))
	kubectlCmd.Stdout = os.Stdout
	kubectlCmd.Stderr = os.Stderr
	kubectlCmd.Stdin = os.Stdin

	if err := kubectlCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("failed to execute kubectl: %w", err)
	}

	return nil
}

// getOrFetchKubeconfig retrieves the kubeconfig from cache or fetches it from the API
func getOrFetchKubeconfig(tenantAPI *api.TenantAPI, tenantID string, noCache bool) (string, error) {
	// Create cache directory
	cacheDir := filepath.Join(os.TempDir(), "spacectl-kubeconfigs")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Generate cache filename using tenant ID hash
	hash := md5.Sum([]byte(tenantID))
	cacheFile := filepath.Join(cacheDir, hex.EncodeToString(hash[:])+".yaml")

	// Check if cached file exists and is fresh (less than 1 hour old)
	if !noCache {
		if info, err := os.Stat(cacheFile); err == nil {
			age := time.Since(info.ModTime())
			if age < 1*time.Hour {
				if debug {
					fmt.Fprintf(os.Stderr, "Using cached kubeconfig (age: %s)\n", age.Round(time.Second))
				}
				return cacheFile, nil
			}
			if debug {
				fmt.Fprintf(os.Stderr, "Cache expired (age: %s), fetching fresh kubeconfig\n", age.Round(time.Second))
			}
		}
	} else if debug {
		fmt.Fprintln(os.Stderr, "Cache disabled, fetching fresh kubeconfig")
	}

	// Fetch kubeconfig from API
	if debug {
		fmt.Fprintf(os.Stderr, "Fetching kubeconfig for tenant %s...\n", tenantID)
	}

	kubeconfig, err := tenantAPI.GetTenantKubeconfig(tenantID)
	if err != nil {
		return "", err
	}

	// Write to cache file
	if err := os.WriteFile(cacheFile, []byte(kubeconfig), 0600); err != nil {
		return "", fmt.Errorf("failed to write kubeconfig to cache: %w", err)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Kubeconfig cached at %s\n", cacheFile)
	}

	return cacheFile, nil
}
