package api

import (
	"fmt"

	"spacectl/internal/models"
)

// OrganizationAPI handles organization-related API calls
type OrganizationAPI struct {
	client *Client
}

// NewOrganizationAPI creates a new OrganizationAPI
func NewOrganizationAPI(client *Client) *OrganizationAPI {
	return &OrganizationAPI{client: client}
}

// ListUserOrganizations lists organizations the user belongs to
func (o *OrganizationAPI) ListUserOrganizations() ([]models.OrganizationMembershipResponse, error) {
	resp, err := o.client.doRequest("GET", "/api/v1/organizations", nil)
	if err != nil {
		return nil, err
	}

	var orgs []models.OrganizationMembershipResponse
	if err := o.client.handleResponse(resp, &orgs); err != nil {
		return nil, err
	}

	return orgs, nil
}

// GetDefaultOrganization gets the user's default organization
func (o *OrganizationAPI) GetDefaultOrganization() (*models.Organization, error) {
	resp, err := o.client.doRequest("GET", "/api/v1/organizations/default", nil)
	if err != nil {
		return nil, err
	}

	var org models.Organization
	if err := o.client.handleResponse(resp, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// GetOrganizationByName gets an organization by name
func (o *OrganizationAPI) GetOrganizationByName(name string) (*models.Organization, error) {
	resp, err := o.client.doRequest("GET", fmt.Sprintf("/api/v1/organizations/by-name/%s", name), nil)
	if err != nil {
		return nil, err
	}

	var org models.Organization
	if err := o.client.handleResponse(resp, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// GetOrganization gets an organization by ID
func (o *OrganizationAPI) GetOrganization(id string) (*models.Organization, error) {
	resp, err := o.client.doRequest("GET", fmt.Sprintf("/api/v1/organizations/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var org models.Organization
	if err := o.client.handleResponse(resp, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// CreateOrganization creates a new organization
func (o *OrganizationAPI) CreateOrganization(name string) (*models.Organization, error) {
	req := models.CreateOrganizationRequest{
		Name: name,
	}

	resp, err := o.client.doRequest("POST", "/api/v1/organizations", req)
	if err != nil {
		return nil, err
	}

	var org models.Organization
	if err := o.client.handleResponse(resp, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// UpdateOrganization updates an organization
func (o *OrganizationAPI) UpdateOrganization(id, name string) (*models.Organization, error) {
	req := models.UpdateOrganizationRequest{
		Name: name,
	}

	resp, err := o.client.doRequest("PUT", fmt.Sprintf("/api/v1/organizations/%s", id), req)
	if err != nil {
		return nil, err
	}

	var org models.Organization
	if err := o.client.handleResponse(resp, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// DeleteOrganization deletes an organization
func (o *OrganizationAPI) DeleteOrganization(id string) error {
	resp, err := o.client.doRequest("DELETE", fmt.Sprintf("/api/v1/organizations/%s", id), nil)
	if err != nil {
		return err
	}

	return o.client.handleResponse(resp, nil)
}

// SetDefaultOrganization sets an organization as default
func (o *OrganizationAPI) SetDefaultOrganization(id string) error {
	resp, err := o.client.doRequest("PUT", fmt.Sprintf("/api/v1/organizations/%s/default", id), nil)
	if err != nil {
		return err
	}

	return o.client.handleResponse(resp, nil)
}

// AddUserToOrganization adds a user to an organization
func (o *OrganizationAPI) AddUserToOrganization(orgID, userID, role string) error {
	req := models.AddUserToOrganizationRequest{
		UserID: userID,
		Role:   role,
	}

	resp, err := o.client.doRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/users", orgID), req)
	if err != nil {
		return err
	}

	return o.client.handleResponse(resp, nil)
}

// RemoveUserFromOrganization removes a user from an organization
func (o *OrganizationAPI) RemoveUserFromOrganization(orgID, userID string) error {
	resp, err := o.client.doRequest("DELETE", fmt.Sprintf("/api/v1/organizations/%s/users/%s", orgID, userID), nil)
	if err != nil {
		return err
	}

	return o.client.handleResponse(resp, nil)
}

// ChangeUserRole changes a user's role in an organization
func (o *OrganizationAPI) ChangeUserRole(orgID, userID, role string) error {
	req := models.ChangeUserRoleRequest{
		Role: role,
	}

	resp, err := o.client.doRequest("PATCH", fmt.Sprintf("/api/v1/organizations/%s/users/%s/role", orgID, userID), req)
	if err != nil {
		return err
	}

	return o.client.handleResponse(resp, nil)
}

// SendInvitation sends an organization invitation
func (o *OrganizationAPI) SendInvitation(orgID, email, role string) error {
	req := models.CreateInvitationRequest{
		Email: email,
		Role:  role,
	}

	resp, err := o.client.doRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/invitations", orgID), req)
	if err != nil {
		return err
	}

	return o.client.handleResponse(resp, nil)
}

// ListOrganizationInvitations lists invitations sent by an organization
func (o *OrganizationAPI) ListOrganizationInvitations(orgID string) ([]models.Invitation, error) {
	resp, err := o.client.doRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/invitations", orgID), nil)
	if err != nil {
		return nil, err
	}

	var invitations []models.Invitation
	if err := o.client.handleResponse(resp, &invitations); err != nil {
		return nil, err
	}

	return invitations, nil
}

// ListUserInvitations lists invitations for the current user
func (o *OrganizationAPI) ListUserInvitations() ([]models.Invitation, error) {
	resp, err := o.client.doRequest("GET", "/api/v1/organizations/invitations", nil)
	if err != nil {
		return nil, err
	}

	var invitations []models.Invitation
	if err := o.client.handleResponse(resp, &invitations); err != nil {
		return nil, err
	}

	return invitations, nil
}

// AcceptInvitation accepts an organization invitation
func (o *OrganizationAPI) AcceptInvitation(invitationID string) error {
	resp, err := o.client.doRequest("POST", fmt.Sprintf("/api/v1/organizations/invitations/%s/accept", invitationID), nil)
	if err != nil {
		return err
	}

	return o.client.handleResponse(resp, nil)
}

// DeclineInvitation declines an organization invitation
func (o *OrganizationAPI) DeclineInvitation(invitationID string) error {
	resp, err := o.client.doRequest("POST", fmt.Sprintf("/api/v1/organizations/invitations/%s/decline", invitationID), nil)
	if err != nil {
		return err
	}

	return o.client.handleResponse(resp, nil)
}
