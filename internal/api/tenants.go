package api

import (
	"fmt"
	"io"
	"net/http"

	"spacectl/internal/models"
)

// TenantAPI handles tenant-related API calls
type TenantAPI struct {
	client *Client
}

// NewTenantAPI creates a new TenantAPI
func NewTenantAPI(client *Client) *TenantAPI {
	return &TenantAPI{client: client}
}

// ListProjectTenants lists tenants in a project
func (t *TenantAPI) ListProjectTenants(projectID string) ([]models.Tenant, error) {
	resp, err := t.client.doRequest("GET", fmt.Sprintf("/api/v1/projects/%s/tenants", projectID), nil)
	if err != nil {
		return nil, err
	}

	var tenants []models.Tenant
	if err := t.client.handleResponse(resp, &tenants); err != nil {
		return nil, err
	}

	return tenants, nil
}

// GetTenant gets a tenant by ID
func (t *TenantAPI) GetTenant(id string) (*models.Tenant, error) {
	resp, err := t.client.doRequest("GET", fmt.Sprintf("/api/v1/tenants/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var tenant models.Tenant
	if err := t.client.handleResponse(resp, &tenant); err != nil {
		return nil, err
	}

	return &tenant, nil
}

// CreateTenant creates a new tenant
func (t *TenantAPI) CreateTenant(projectID string, req models.CreateTenantRequest) (*models.Tenant, error) {
	resp, err := t.client.doRequest("POST", fmt.Sprintf("/api/v1/projects/%s/tenants", projectID), req)
	if err != nil {
		return nil, err
	}

	var tenant models.Tenant
	if err := t.client.handleResponse(resp, &tenant); err != nil {
		return nil, err
	}

	return &tenant, nil
}

// UpdateTenant updates a tenant
func (t *TenantAPI) UpdateTenant(id string, req models.UpdateTenantRequest) (*models.Tenant, error) {
	resp, err := t.client.doRequest("PATCH", fmt.Sprintf("/api/v1/tenants/%s", id), req)
	if err != nil {
		return nil, err
	}

	var tenant models.Tenant
	if err := t.client.handleResponse(resp, &tenant); err != nil {
		return nil, err
	}

	return &tenant, nil
}

// DeleteTenant deletes a tenant
func (t *TenantAPI) DeleteTenant(id string) error {
	resp, err := t.client.doRequest("DELETE", fmt.Sprintf("/api/v1/tenants/%s", id), nil)
	if err != nil {
		return err
	}

	return t.client.handleResponse(resp, nil)
}

// GetTenantStatus gets tenant provisioning status
func (t *TenantAPI) GetTenantStatus(id string) (*models.TenantStatusResponse, error) {
	resp, err := t.client.doRequest("GET", fmt.Sprintf("/api/v1/tenants/%s/status", id), nil)
	if err != nil {
		return nil, err
	}

	var status models.TenantStatusResponse
	if err := t.client.handleResponse(resp, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// GetTenantKubeconfig gets tenant kubeconfig
func (t *TenantAPI) GetTenantKubeconfig(id string) (string, error) {
	resp, err := t.client.doRequest("GET", fmt.Sprintf("/api/v1/tenants/%s/kubeconfig", id), nil)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get kubeconfig: status %d", resp.StatusCode)
	}

	// Read the response body as string (kubeconfig content)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read kubeconfig response: %w", err)
	}

	return string(body), nil
}

// GetAvailableLocations gets available cloud locations
func (t *TenantAPI) GetAvailableLocations() ([]models.Location, error) {
	resp, err := t.client.doRequest("GET", "/api/v1/tenants/locations", nil)
	if err != nil {
		return nil, err
	}

	var locations []models.Location
	if err := t.client.handleResponse(resp, &locations); err != nil {
		return nil, err
	}

	return locations, nil
}

// GetAvailableClouds gets available cloud providers
func (t *TenantAPI) GetAvailableClouds() ([]string, error) {
	resp, err := t.client.doRequest("GET", "/api/v1/tenants/clouds", nil)
	if err != nil {
		return nil, err
	}

	var clouds []string
	if err := t.client.handleResponse(resp, &clouds); err != nil {
		return nil, err
	}

	return clouds, nil
}

// GetAvailableRegions gets available regions for a cloud provider
func (t *TenantAPI) GetAvailableRegions(cloudProvider string) ([]string, error) {
	resp, err := t.client.doRequest("GET", fmt.Sprintf("/api/v1/tenants/regions?cloud_provider=%s", cloudProvider), nil)
	if err != nil {
		return nil, err
	}

	var regions []string
	if err := t.client.handleResponse(resp, &regions); err != nil {
		return nil, err
	}

	return regions, nil
}

// GetAvailableZones gets available zones for a cloud provider and region
func (t *TenantAPI) GetAvailableZones(cloudProvider, region string) ([]string, error) {
	resp, err := t.client.doRequest("GET", fmt.Sprintf("/api/v1/tenants/zones?cloud_provider=%s&region=%s", cloudProvider, region), nil)
	if err != nil {
		return nil, err
	}

	var zones []string
	if err := t.client.handleResponse(resp, &zones); err != nil {
		return nil, err
	}

	return zones, nil
}

// GetAvailableKubernetesVersions gets available Kubernetes versions
func (t *TenantAPI) GetAvailableKubernetesVersions() ([]models.KubernetesVersion, error) {
	resp, err := t.client.doRequest("GET", "/api/v1/tenants/kubernetes-versions", nil)
	if err != nil {
		return nil, err
	}

	var versions []models.KubernetesVersion
	if err := t.client.handleResponse(resp, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}
