package cmd

import (
	"fmt"

	"spacectl/internal/api"
)

// resolveOrganizationID resolves an organization identifier from either name or id.
// If both are empty, returns an error. If both are provided, returns an error.
func resolveOrganizationID(client *api.Client, name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("either --name or --id must be provided")
	}
	if name != "" && id != "" {
		return "", fmt.Errorf("only one of --name or --id is allowed")
	}
	if id != "" {
		return id, nil
	}
	orgAPI := api.NewOrganizationAPI(client)
	org, err := orgAPI.GetOrganizationByName(name)
	if err != nil {
		return "", fmt.Errorf("failed to resolve organization by name: %w", err)
	}
	return org.ID, nil
}

// resolveProjectID resolves a project ID from name or id, optionally within an organization.
// If orgID is provided, the search is scoped; otherwise falls back to the user's projects.
func resolveProjectID(client *api.Client, projectName, projectID, orgID string) (string, error) {
	if projectName == "" && projectID == "" {
		return "", fmt.Errorf("either --name or --id must be provided for project")
	}
	if projectName != "" && projectID != "" {
		return "", fmt.Errorf("only one of --name or --id is allowed for project")
	}
	if projectID != "" {
		return projectID, nil
	}
	projectAPI := api.NewProjectAPI(client)
	if orgID != "" {
		projects, err := projectAPI.ListOrganizationProjects(orgID)
		if err != nil {
			return "", fmt.Errorf("failed to list projects in organization: %w", err)
		}
		for _, p := range projects {
			if p.Name == projectName {
				return p.ID, nil
			}
		}
		return "", fmt.Errorf("project named %q not found in organization", projectName)
	}
	// Fallback: search user's projects
	memberships, err := projectAPI.ListUserProjects()
	if err != nil {
		return "", fmt.Errorf("failed to list user projects: %w", err)
	}
	for _, m := range memberships {
		if m.Project.Name == projectName {
			return m.Project.ID, nil
		}
	}
	return "", fmt.Errorf("project named %q not found", projectName)
}

// resolveTenantID resolves a tenant ID from name or id within a project.
func resolveTenantID(client *api.Client, tenantName, tenantID, projectID string) (string, error) {
	if tenantName == "" && tenantID == "" {
		return "", fmt.Errorf("either --name or --id must be provided for tenant")
	}
	if tenantName != "" && tenantID != "" {
		return "", fmt.Errorf("only one of --name or --id is allowed for tenant")
	}
	if tenantID != "" {
		return tenantID, nil
	}
	if projectID == "" {
		return "", fmt.Errorf("project is required to resolve tenant by name")
	}
	tenantAPI := api.NewTenantAPI(client)
	tenants, err := tenantAPI.ListProjectTenants(projectID)
	if err != nil {
		return "", fmt.Errorf("failed to list tenants in project: %w", err)
	}
	for _, t := range tenants {
		if t.Name == tenantName {
			return t.ID, nil
		}
	}
	return "", fmt.Errorf("tenant with name %q not found in project", tenantName)
}
