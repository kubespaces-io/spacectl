package cmd

import (
	"fmt"

	"spacectl/internal/api"
	"spacectl/internal/models"

	"github.com/spf13/cobra"
)

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  `Manage projects including listing, creating, updating, and deleting them.`,
}

func init() {
	rootCmd.AddCommand(projectCmd)
}

// projectListCmd represents the project list command
var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Long:  `List projects. Use --org to filter by organization.`,
	RunE:  runProjectList,
}

var projectListOrg string
var projectListOrgName string
var projectListAll bool

func init() {
	projectCmd.AddCommand(projectListCmd)
	projectListCmd.Flags().StringVar(&projectListOrg, "org", "", "Organization ID to filter projects")
	projectListCmd.Flags().StringVar(&projectListOrgName, "org-name", "", "Organization name to filter projects")
	projectListCmd.Flags().BoolVar(&projectListAll, "all", false, "List projects from all organizations")
}

func runProjectList(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl auth login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	projectAPI := api.NewProjectAPI(client)
	orgAPI := api.NewOrganizationAPI(client)
	tenantAPI := api.NewTenantAPI(client)

	// Validate flags
	if projectListOrgName != "" && projectListOrg != "" {
		return fmt.Errorf("only one of --org or --org-name is allowed")
	}
	if projectListAll && (projectListOrg != "" || projectListOrgName != "") {
		return fmt.Errorf("--all cannot be used with --org or --org-name")
	}

	if projectListAll {
		// List projects from all organizations with tenant counts
		return runProjectListAll(client, projectAPI, orgAPI, tenantAPI)
	}

	// Determine target organization
	var targetOrgID string
	if projectListOrgName != "" {
		org, err := orgAPI.GetOrganizationByName(projectListOrgName)
		if err != nil {
			return fmt.Errorf("failed to resolve organization by name: %w", err)
		}
		targetOrgID = org.ID
	} else if projectListOrg != "" {
		targetOrgID = projectListOrg
	} else {
		// Use default organization
		defOrg, err := orgAPI.GetDefaultOrganization()
		if err != nil {
			return fmt.Errorf("failed to get default organization: %w", err)
		}
		targetOrgID = defOrg.ID
	}

	// List projects in target organization with tenant counts
	return runProjectListForOrg(client, projectAPI, tenantAPI, targetOrgID)
}

// runProjectListForOrg lists projects in a specific organization with tenant counts
func runProjectListForOrg(client *api.Client, projectAPI *api.ProjectAPI, tenantAPI *api.TenantAPI, orgID string) error {
	// Get projects in organization
	projects, err := projectAPI.ListOrganizationProjects(orgID)
	if err != nil {
		return fmt.Errorf("failed to list organization projects: %w", err)
	}

	// Create enhanced project list with tenant counts
	var enhancedProjects []map[string]interface{}
	for _, project := range projects {
		// Get tenant count for this project
		tenants, err := tenantAPI.ListProjectTenants(project.ID)
		if err != nil {
			// If we can't get tenant count, continue with 0
			tenants = []models.Tenant{}
		}

		enhancedProject := map[string]interface{}{
			"name":         project.Name,
			"role":         "admin", // Default role for org projects
			"tenant_count": len(tenants),
		}
		enhancedProjects = append(enhancedProjects, enhancedProject)
	}

	return formatter.FormatData(enhancedProjects)
}

// runProjectListAll lists projects from all organizations with tenant counts
func runProjectListAll(client *api.Client, projectAPI *api.ProjectAPI, orgAPI *api.OrganizationAPI, tenantAPI *api.TenantAPI) error {
	// Get all user organizations
	orgs, err := orgAPI.ListUserOrganizations()
	if err != nil {
		return fmt.Errorf("failed to list user organizations: %w", err)
	}

	// Collect all projects with tenant counts
	var allProjects []map[string]interface{}
	for _, orgMembership := range orgs {
		projects, err := projectAPI.ListOrganizationProjects(orgMembership.Organization.ID)
		if err != nil {
			// Skip organizations where we can't list projects
			continue
		}

		for _, project := range projects {
			// Get tenant count for this project
			tenants, err := tenantAPI.ListProjectTenants(project.ID)
			if err != nil {
				// If we can't get tenant count, continue with 0
				tenants = []models.Tenant{}
			}

			enhancedProject := map[string]interface{}{
				"organization": orgMembership.Organization.Name,
				"name":         project.Name,
				"role":         orgMembership.Role,
				"tenant_count": len(tenants),
			}
			allProjects = append(allProjects, enhancedProject)
		}
	}

	return formatter.FormatData(allProjects)
}

// projectCreateCmd represents the project create command
var projectCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a project",
	Long:  `Create a new project in the specified organization.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectCreate,
}

var (
	projectCreateOrg        string
	projectCreateOrgName    string
	projectCreateDesc       string
	projectCreateMaxTenants int
	projectCreateMaxCompute int
	projectCreateMaxMemory  int
)

func init() {
	projectCmd.AddCommand(projectCreateCmd)
	projectCreateCmd.Flags().StringVar(&projectCreateOrg, "org", "", "Organization ID")
	projectCreateCmd.Flags().StringVar(&projectCreateOrgName, "org-name", "", "Organization name")
	projectCreateCmd.Flags().StringVar(&projectCreateDesc, "description", "", "Project description")
	projectCreateCmd.Flags().IntVar(&projectCreateMaxTenants, "max-tenants", 0, "Maximum number of tenants")
	projectCreateCmd.Flags().IntVar(&projectCreateMaxCompute, "max-compute", 0, "Maximum compute quota")
	projectCreateCmd.Flags().IntVar(&projectCreateMaxMemory, "max-memory", 0, "Maximum memory quota (GB)")
}

func runProjectCreate(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	name := args[0]

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	projectAPI := api.NewProjectAPI(client)
	orgAPI := api.NewOrganizationAPI(client)

	// Resolve org if name used
	if projectCreateOrgName != "" && projectCreateOrg != "" {
		return fmt.Errorf("only one of --org or --org-name is allowed")
	}
	if projectCreateOrg == "" && projectCreateOrgName != "" {
		org, err := orgAPI.GetOrganizationByName(projectCreateOrgName)
		if err != nil {
			return fmt.Errorf("failed to resolve organization by name: %w", err)
		}
		projectCreateOrg = org.ID
	}
	// If still empty, use default organization
	if projectCreateOrg == "" {
		def, err := orgAPI.GetDefaultOrganization()
		if err != nil {
			return fmt.Errorf("failed to get default organization: %w", err)
		}
		projectCreateOrg = def.ID
	}

	// Prepare request
	req := models.CreateProjectRequest{
		Name:        name,
		MaxTenants:  projectCreateMaxTenants,
		MaxCompute:  projectCreateMaxCompute,
		MaxMemoryGB: projectCreateMaxMemory,
	}

	if projectCreateDesc != "" {
		req.Description = &projectCreateDesc
	}

	// Create project
	project, err := projectAPI.CreateProject(projectCreateOrg, req)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Output project
	return formatter.FormatData(project)
}

// projectGetCmd represents the project get command
var projectGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get project details",
	Long:  `Get detailed information about a specific project.`,
	Args:  cobra.NoArgs,
	RunE:  runProjectGet,
}

func init() {
	projectCmd.AddCommand(projectGetCmd)
	projectGetCmd.Flags().StringVar(&projectGetID, "project-id", "", "Project ID")
	projectGetCmd.Flags().StringVar(&projectGetName, "project-name", "", "Project name")
}

var (
	projectGetID   string
	projectGetName string
)

func runProjectGet(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	projectAPI := api.NewProjectAPI(client)

	// Resolve project
	if projectGetID != "" && projectGetName != "" {
		return fmt.Errorf("only one of --project-id or --project-name is allowed")
	}
	id := projectGetID
	if id == "" {
		var err error
		id, err = resolveProjectID(client, projectGetName, "", "")
		if err != nil {
			return err
		}
	}

	// Get project
	project, err := projectAPI.GetProject(id)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Output project
	return formatter.FormatData(project)
}

// projectUpdateCmd represents the project update command
var projectUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a project",
	Long:  `Update a project's metadata.`,
	Args:  cobra.NoArgs,
	RunE:  runProjectUpdate,
}

var (
	projectUpdateName       string
	projectUpdateDesc       string
	projectUpdateMaxTenants int
	projectUpdateMaxCompute int
	projectUpdateMaxMemory  int
	projectUpdateTargetID   string
	projectUpdateTargetName string
)

func init() {
	projectCmd.AddCommand(projectUpdateCmd)
	projectUpdateCmd.Flags().StringVar(&projectUpdateName, "name", "", "New project name")
	projectUpdateCmd.Flags().StringVar(&projectUpdateDesc, "description", "", "New project description")
	projectUpdateCmd.Flags().IntVar(&projectUpdateMaxTenants, "max-tenants", -1, "New maximum number of tenants")
	projectUpdateCmd.Flags().IntVar(&projectUpdateMaxCompute, "max-compute", -1, "New maximum compute quota")
	projectUpdateCmd.Flags().IntVar(&projectUpdateMaxMemory, "max-memory", -1, "New maximum memory quota (GB)")
	projectUpdateCmd.Flags().StringVar(&projectUpdateTargetID, "project-id", "", "Project ID to update")
	projectUpdateCmd.Flags().StringVar(&projectUpdateTargetName, "project-name", "", "Project name to update")
}

func runProjectUpdate(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	projectAPI := api.NewProjectAPI(client)

	// Resolve target project by name or id
	if projectUpdateTargetID != "" && projectUpdateTargetName != "" {
		return fmt.Errorf("only one of --project-id or --project-name is allowed")
	}
	id := projectUpdateTargetID
	if id == "" {
		var err error
		id, err = resolveProjectID(client, projectUpdateTargetName, "", "")
		if err != nil {
			return err
		}
	}

	// Get current project to fill in missing fields
	currentProject, err := projectAPI.GetProject(id)
	if err != nil {
		return fmt.Errorf("failed to get current project: %w", err)
	}

	// Prepare request
	req := models.UpdateProjectRequest{
		Name:        projectUpdateName,
		Description: &projectUpdateDesc,
		MaxTenants:  projectUpdateMaxTenants,
		MaxCompute:  projectUpdateMaxCompute,
		MaxMemoryGB: projectUpdateMaxMemory,
	}

	// Use current values for fields not provided
	if req.Name == "" {
		req.Name = currentProject.Name
	}
	if req.Description == nil || *req.Description == "" {
		req.Description = currentProject.Description
	}
	if req.MaxTenants == -1 {
		req.MaxTenants = currentProject.MaxTenants
	}
	if req.MaxCompute == -1 {
		req.MaxCompute = currentProject.MaxCompute
	}
	if req.MaxMemoryGB == -1 {
		req.MaxMemoryGB = currentProject.MaxMemoryGB
	}

	// Update project
	project, err := projectAPI.UpdateProject(id, req)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	// Output project
	return formatter.FormatData(project)
}

// projectDeleteCmd represents the project delete command
var projectDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a project",
	Long:  `Delete a project. This action cannot be undone.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectDelete,
}

func init() {
	projectCmd.AddCommand(projectDeleteCmd)
}

func runProjectDelete(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	id := args[0]

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	projectAPI := api.NewProjectAPI(client)

	// Delete project
	err := projectAPI.DeleteProject(id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Output success message
	if !quiet {
		fmt.Printf("Successfully deleted project %s\n", id)
	}

	return nil
}

// projectMembersCmd represents the project members command
var projectMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "Manage project members",
	Long:  `Manage project members including listing, adding, and removing them.`,
}

func init() {
	projectCmd.AddCommand(projectMembersCmd)
}

// projectMembersListCmd represents the project members list command
var projectMembersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List project members",
	Long:  `List all members of a project.`,
	Args:  cobra.NoArgs,
	RunE:  runProjectMembersList,
}

func init() {
	projectMembersCmd.AddCommand(projectMembersListCmd)
	projectMembersListCmd.Flags().StringVar(&projectMembersListProjID, "project-id", "", "Project ID")
	projectMembersListCmd.Flags().StringVar(&projectMembersListProjName, "project-name", "", "Project name")
}

var (
	projectMembersListProjID   string
	projectMembersListProjName string
)

func runProjectMembersList(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	// Resolve project
	projectID, err := resolveProjectID(client, projectMembersListProjName, projectMembersListProjID, "")
	if err != nil {
		return err
	}
	projectAPI := api.NewProjectAPI(client)

	// Get project members
	members, err := projectAPI.ListProjectMembers(projectID)
	if err != nil {
		return fmt.Errorf("failed to list project members: %w", err)
	}

	// Output members
	return formatter.FormatData(members)
}

// projectMembersAddCmd represents the project members add command
var projectMembersAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a member to a project",
	Long:  `Add a user to a project with the specified role.`,
	Args:  cobra.NoArgs,
	RunE:  runProjectMembersAdd,
}

var (
	projectMembersAddUserID   string
	projectMembersAddRole     string
	projectMembersAddProjID   string
	projectMembersAddProjName string
)

func init() {
	projectMembersCmd.AddCommand(projectMembersAddCmd)
	projectMembersAddCmd.Flags().StringVar(&projectMembersAddUserID, "user", "", "User ID to add")
	projectMembersAddCmd.Flags().StringVar(&projectMembersAddRole, "role", "", "Role (admin, member)")
	projectMembersAddCmd.Flags().StringVar(&projectMembersAddProjID, "project-id", "", "Project ID")
	projectMembersAddCmd.Flags().StringVar(&projectMembersAddProjName, "project-name", "", "Project name")
	projectMembersAddCmd.MarkFlagRequired("user")
	projectMembersAddCmd.MarkFlagRequired("role")
}

func runProjectMembersAdd(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	// Resolve project
	projectID, err := resolveProjectID(client, projectMembersAddProjName, projectMembersAddProjID, "")
	if err != nil {
		return err
	}

	projectAPI := api.NewProjectAPI(client)

	// Add user to project
	err = projectAPI.AddUserToProject(projectID, projectMembersAddUserID, projectMembersAddRole)
	if err != nil {
		return fmt.Errorf("failed to add user to project: %w", err)
	}

	// Output success message
	if !quiet {
		fmt.Printf("Successfully added user %s to project %s with role %s\n",
			projectMembersAddUserID, projectID, projectMembersAddRole)
	}

	return nil
}

// projectMembersRemoveCmd represents the project members remove command
var projectMembersRemoveCmd = &cobra.Command{
	Use:   "remove <project-id> <user-id>",
	Short: "Remove a member from a project",
	Long:  `Remove a user from a project.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runProjectMembersRemove,
}

func init() {
	projectMembersCmd.AddCommand(projectMembersRemoveCmd)
}

func runProjectMembersRemove(cmd *cobra.Command, args []string) error {
	// Check if user is authenticated
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'spacectl login' first")
	}

	projectID := args[0]
	userID := args[1]

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg, debug)
	projectAPI := api.NewProjectAPI(client)

	// Remove user from project
	err := projectAPI.RemoveUserFromProject(projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from project: %w", err)
	}

	// Output success message
	if !quiet {
		fmt.Printf("Successfully removed user %s from project %s\n", userID, projectID)
	}

	return nil
}
