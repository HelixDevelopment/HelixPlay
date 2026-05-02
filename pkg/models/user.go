package models

import "time"

// User represents a player with OAuth2/OIDC identity, scoped to a Tenant
type User struct {
	ID                     string    `json:"id" db:"id"`
	TenantID               string    `json:"tenant_id" db:"tenant_id"`
	Email                  string    `json:"email" db:"email"`
	DisplayName            string    `json:"display_name" db:"display_name"`
	Roles                  []string  `json:"roles" db:"roles"`
	OAuth2Subject          string    `json:"oauth2_subject" db:"oauth2_subject"`
	OAuth2Provider         string    `json:"oauth2_provider" db:"oauth2_provider"`
	PreferredControllerType string   `json:"preferred_controller_type" db:"preferred_controller_type"`
	StorageQuotaUsedGB     float64   `json:"storage_quota_used_gb" db:"storage_quota_used_gb"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	LastLoginAt            *time.Time `json:"last_login_at" db:"last_login_at"`
}
