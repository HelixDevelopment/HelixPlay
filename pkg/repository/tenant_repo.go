package repository

import (
	"context"
	"github.com/HelixDevelopment/HelixPlay/pkg/models"
)

// TenantRepository defines the interface for tenant data access
type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	GetByID(ctx context.Context, id string) (*models.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	Update(ctx context.Context, tenant *models.Tenant) error
	UpdateTheme(ctx context.Context, id string, primaryColor, secondaryColor, logoURL string) error
	UpdateCatalogFilter(ctx context.Context, id string, filter map[string]any) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Tenant, error)
}
