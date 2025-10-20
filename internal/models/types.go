package models

import "time"

// User represents a user in the system
type User struct {
	ID            string           `json:"id"`
	Email         string           `json:"email"`
	Provider      string           `json:"provider"`
	Approved      bool             `json:"approved"`
	EmailVerified bool             `json:"email_verified"`
	IsAdmin       bool             `json:"is_admin"`
	Preferences   *UserPreferences `json:"preferences"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

type UserPreferences struct {
	WelcomeDismissed bool   `json:"welcome_dismissed"`
	Theme            string `json:"theme"`
}

// Organization represents an organization
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserOrganization struct {
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	Role           string    `json:"role"`
	IsDefault      bool      `json:"is_default"`
	CreatedAt      time.Time `json:"created_at"`
}

type OrganizationMembershipResponse struct {
	Organization Organization `json:"organization"`
	Role         string       `json:"role"`
	IsDefault    bool         `json:"is_default"`
}

// Project represents a project
type Project struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	MaxTenants     int       `json:"max_tenants"`
	MaxCompute     int       `json:"max_compute"`
	MaxMemoryGB    int       `json:"max_memory_gb"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ProjectMembership struct {
	Project   Project   `json:"project"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type ProjectMember struct {
	UserID    string    `json:"user_id"`
	ProjectID string    `json:"project_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Tenant represents a Kubernetes tenant
type Tenant struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	OrganizationID    string    `json:"organization_id"`
	HostClusterID     string    `json:"host_cluster_id"`
	Name              string    `json:"name"`
	CloudProvider     string    `json:"cloud_provider"`
	Region            string    `json:"region"`
	LocationShort     string    `json:"location_short"`
	KubernetesVersion string    `json:"kubernetes_version"`
	ComputeQuota      int       `json:"compute_quota"`
	MemoryQuotaGB     int       `json:"memory_quota_gb"`
	Status            string    `json:"status"`
	Namespace         string    `json:"namespace"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type TenantStatusResponse struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Status            string    `json:"status"`
	Namespace         string    `json:"namespace"`
	CloudProvider     string    `json:"cloud_provider"`
	Region            string    `json:"region"`
	KubernetesVersion string    `json:"kubernetes_version"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Invitation represents an organization invitation
type Invitation struct {
	ID            string       `json:"id"`
	Organization  Organization `json:"organization"`
	InviterUserID string       `json:"inviter_user_id"`
	InviteeEmail  string       `json:"invitee_email"`
	Role          string       `json:"role"`
	Status        string       `json:"status"`
	ExpiresAt     time.Time    `json:"expires_at"`
	CreatedAt     time.Time    `json:"created_at"`
}

// ProjectInvitation represents a project invitation
type ProjectInvitation struct {
	ID             string    `json:"id"`
	Project        Project   `json:"project"`
	OrganizationID string    `json:"organization_id"`
	InviterUserID  string    `json:"inviter_user_id"`
	InviteeEmail   string    `json:"invitee_email"`
	Role           string    `json:"role"`
	Status         string    `json:"status"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// KubernetesVersion represents an available Kubernetes version
type KubernetesVersion struct {
	Version   string `json:"version"`
	IsDefault bool   `json:"is_default"`
}

// Location represents a cloud location
type Location struct {
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
	Zone          string `json:"zone"`
}

// Auth types
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// Request/Response types for CRUD operations
type CreateOrganizationRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UpdateOrganizationRequest struct {
	Name string `json:"name"`
}

type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	MaxTenants  int     `json:"max_tenants"`
	MaxCompute  int     `json:"max_compute"`
	MaxMemoryGB int     `json:"max_memory_gb"`
}

type UpdateProjectQuotasRequest struct {
	MaxTenants  int `json:"max_tenants"`
	MaxCompute  int `json:"max_compute"`
	MaxMemoryGB int `json:"max_memory_gb"`
}

type UpdateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	MaxTenants  int     `json:"max_tenants"`
	MaxCompute  int     `json:"max_compute"`
	MaxMemoryGB int     `json:"max_memory_gb"`
}

type CreateTenantRequest struct {
	Name              string `json:"name"`
	CloudProvider     string `json:"cloud_provider"`
	Region            string `json:"region"`
	KubernetesVersion string `json:"kubernetes_version"`
	ComputeQuota      int    `json:"compute_quota"`
	MemoryQuotaGB     int    `json:"memory_quota_gb"`
	NamespaceSuffix   string `json:"namespace_suffix"`
}

type UpdateTenantRequest struct {
	KubernetesVersion *string `json:"kubernetes_version"`
	ComputeQuota      *int    `json:"compute_quota"`
	MemoryQuotaGB     *int    `json:"memory_quota_gb"`
}

type AddUserToOrganizationRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type ChangeUserRoleRequest struct {
	Role string `json:"role"`
}

type AddUserToProjectRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type ChangeProjectUserRoleRequest struct {
	Role string `json:"role"`
}

type CreateInvitationRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type CreateProjectInvitationRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Error response
type ErrorResponse struct {
	Error string `json:"error"`
}
