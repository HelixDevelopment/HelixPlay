package models

import "time"

// Tenant represents a white-label partner (ISP, hotel, hospital, venue)
type Tenant struct {
	ID                      string          `json:"id" db:"id"`
	Slug                    string          `json:"slug" db:"slug"`
	Name                    string          `json:"name" db:"name"`
	ThemePrimaryColor       string          `json:"theme_primary_color" db:"theme_primary_color"`
	ThemeSecondaryColor     string          `json:"theme_secondary_color" db:"theme_secondary_color"`
	ThemeLogoURL            string          `json:"theme_logo_url" db:"theme_logo_url"`
	OAuth2Provider          string          `json:"oauth2_provider" db:"oauth2_provider"`
	OAuth2Config            map[string]any  `json:"oauth2_config" db:"oauth2_config"`
	CatalogFilter           map[string]any  `json:"catalog_filter" db:"catalog_filter"`
	MonetizationModel       string          `json:"monetization_model" db:"monetization_model"`
	ResourceQuotaMaxSessions int            `json:"resource_quota_max_sessions" db:"resource_quota_max_sessions"`
	ResourceQuotaStorageGB  int             `json:"resource_quota_storage_gb" db:"resource_quota_storage_gb"`
	CreatedAt               time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at" db:"updated_at"`
}
