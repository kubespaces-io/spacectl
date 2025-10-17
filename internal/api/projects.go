package api

import (
	"fmt"

	"spacectl/internal/models"
)

// ProjectAPI handles project-related API calls
type ProjectAPI struct {
	client *Client
}

// NewProjectAPI creates a new ProjectAPI
func NewProjectAPI(client *Client) *ProjectAPI {
	return &ProjectAPI{client: client}
}

// ListOrganizationProjects lists projects in an organization
func (p *ProjectAPI) ListOrganizationProjects(orgID string) ([]models.Project, error) {
	resp, err := p.client.doRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/projects", orgID), nil)
	if err != nil {
		return nil, err
	}

	var projects []models.Project
	if err := p.client.handleResponse(resp, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// ListUserProjects lists projects the user participates in
func (p *ProjectAPI) ListUserProjects() ([]models.ProjectMembership, error) {
	resp, err := p.client.doRequest("GET", "/api/v1/projects", nil)
	if err != nil {
		return nil, err
	}

	var projects []models.ProjectMembership
	if err := p.client.handleResponse(resp, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// GetProject gets a project by ID
func (p *ProjectAPI) GetProject(id string) (*models.Project, error) {
	resp, err := p.client.doRequest("GET", fmt.Sprintf("/api/v1/projects/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var project models.Project
	if err := p.client.handleResponse(resp, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// CreateProject creates a new project
func (p *ProjectAPI) CreateProject(orgID string, req models.CreateProjectRequest) (*models.Project, error) {
	resp, err := p.client.doRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/projects", orgID), req)
	if err != nil {
		return nil, err
	}

	var project models.Project
	if err := p.client.handleResponse(resp, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// UpdateProject updates a project
func (p *ProjectAPI) UpdateProject(id string, req models.UpdateProjectRequest) (*models.Project, error) {
	resp, err := p.client.doRequest("PUT", fmt.Sprintf("/api/v1/projects/%s", id), req)
	if err != nil {
		return nil, err
	}

	var project models.Project
	if err := p.client.handleResponse(resp, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// UpdateProjectQuotas updates project quotas
func (p *ProjectAPI) UpdateProjectQuotas(id string, req models.UpdateProjectQuotasRequest) (*models.Project, error) {
	resp, err := p.client.doRequest("PATCH", fmt.Sprintf("/api/v1/projects/%s/quotas", id), req)
	if err != nil {
		return nil, err
	}

	var project models.Project
	if err := p.client.handleResponse(resp, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// DeleteProject deletes a project
func (p *ProjectAPI) DeleteProject(id string) error {
	resp, err := p.client.doRequest("DELETE", fmt.Sprintf("/api/v1/projects/%s", id), nil)
	if err != nil {
		return err
	}

	return p.client.handleResponse(resp, nil)
}

// ListProjectMembers lists project members
func (p *ProjectAPI) ListProjectMembers(projectID string) ([]models.ProjectMember, error) {
	resp, err := p.client.doRequest("GET", fmt.Sprintf("/api/v1/projects/%s/users", projectID), nil)
	if err != nil {
		return nil, err
	}

	var members []models.ProjectMember
	if err := p.client.handleResponse(resp, &members); err != nil {
		return nil, err
	}

	return members, nil
}

// AddUserToProject adds a user to a project
func (p *ProjectAPI) AddUserToProject(projectID, userID, role string) error {
	req := models.AddUserToProjectRequest{
		UserID: userID,
		Role:   role,
	}

	resp, err := p.client.doRequest("POST", fmt.Sprintf("/api/v1/projects/%s/users", projectID), req)
	if err != nil {
		return err
	}

	return p.client.handleResponse(resp, nil)
}

// RemoveUserFromProject removes a user from a project
func (p *ProjectAPI) RemoveUserFromProject(projectID, userID string) error {
	resp, err := p.client.doRequest("DELETE", fmt.Sprintf("/api/v1/projects/%s/users/%s", projectID, userID), nil)
	if err != nil {
		return err
	}

	return p.client.handleResponse(resp, nil)
}

// ChangeProjectUserRole changes a user's role in a project
func (p *ProjectAPI) ChangeProjectUserRole(projectID, userID, role string) error {
	req := models.ChangeProjectUserRoleRequest{
		Role: role,
	}

	resp, err := p.client.doRequest("PATCH", fmt.Sprintf("/api/v1/projects/%s/users/%s/role", projectID, userID), req)
	if err != nil {
		return err
	}

	return p.client.handleResponse(resp, nil)
}

// SendProjectInvitation sends a project invitation
func (p *ProjectAPI) SendProjectInvitation(projectID, email, role string) error {
	req := models.CreateProjectInvitationRequest{
		Email: email,
		Role:  role,
	}

	resp, err := p.client.doRequest("POST", fmt.Sprintf("/api/v1/projects/%s/invitations", projectID), req)
	if err != nil {
		return err
	}

	return p.client.handleResponse(resp, nil)
}

// ListProjectInvitations lists invitations sent for a project
func (p *ProjectAPI) ListProjectInvitations(projectID string) ([]models.ProjectInvitation, error) {
	resp, err := p.client.doRequest("GET", fmt.Sprintf("/api/v1/projects/%s/invitations", projectID), nil)
	if err != nil {
		return nil, err
	}

	var invitations []models.ProjectInvitation
	if err := p.client.handleResponse(resp, &invitations); err != nil {
		return nil, err
	}

	return invitations, nil
}

// ListUserProjectInvitations lists project invitations for the current user
func (p *ProjectAPI) ListUserProjectInvitations() ([]models.ProjectInvitation, error) {
	resp, err := p.client.doRequest("GET", "/api/v1/projects/invitations", nil)
	if err != nil {
		return nil, err
	}

	var invitations []models.ProjectInvitation
	if err := p.client.handleResponse(resp, &invitations); err != nil {
		return nil, err
	}

	return invitations, nil
}

// AcceptProjectInvitation accepts a project invitation
func (p *ProjectAPI) AcceptProjectInvitation(invitationID string) error {
	resp, err := p.client.doRequest("POST", fmt.Sprintf("/api/v1/projects/invitations/%s/accept", invitationID), nil)
	if err != nil {
		return err
	}

	return p.client.handleResponse(resp, nil)
}

// DeclineProjectInvitation declines a project invitation
func (p *ProjectAPI) DeclineProjectInvitation(invitationID string) error {
	resp, err := p.client.doRequest("POST", fmt.Sprintf("/api/v1/projects/invitations/%s/decline", invitationID), nil)
	if err != nil {
		return err
	}

	return p.client.handleResponse(resp, nil)
}
