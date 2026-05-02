package tenant

import (
	"context"
	"fmt"
	"time"

	"github.com/HelixDevelopment/HelixPlay/pkg/models"
	"github.com/HelixDevelopment/HelixPlay/pkg/repository"
)

// Service provides CRUD operations for white-label tenants.
type Service struct {
	repo repository.TenantRepository
}

// NewService creates a tenant service backed by the given repository.
func NewService(repo repository.TenantRepository) *Service {
	return &Service{repo: repo}
}

// CreateTenant persists a new tenant after validating required fields.
func (s *Service) CreateTenant(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error) {
	if tenant.Slug == "" {
		return nil, fmt.Errorf("tenant slug is required")
	}
	if tenant.Name == "" {
		return nil, fmt.Errorf("tenant name is required")
	}

	now := time.Now().UTC()
	tenant.ID = generateID()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	return tenant, nil
}

// GetTenant retrieves a tenant by its unique ID.
func (s *Service) GetTenant(ctx context.Context, id string) (*models.Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

// GetTenantBySlug retrieves a tenant by its human-readable slug.
func (s *Service) GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	return s.repo.GetBySlug(ctx, slug)
}

// UpdateTenant applies changes to an existing tenant.
func (s *Service) UpdateTenant(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error) {
	existing, err := s.repo.GetByID(ctx, tenant.ID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	existing.Name = tenant.Name
	existing.Slug = tenant.Slug
	existing.MonetizationModel = tenant.MonetizationModel
	existing.ResourceQuotaMaxSessions = tenant.ResourceQuotaMaxSessions
	existing.ResourceQuotaStorageGB = tenant.ResourceQuotaStorageGB
	existing.OAuth2Provider = tenant.OAuth2Provider
	existing.OAuth2Config = tenant.OAuth2Config
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}
	return existing, nil
}

// UpdateTheme changes the visual branding for a tenant.
func (s *Service) UpdateTheme(ctx context.Context, id, primaryColor, secondaryColor, logoURL string) error {
	return s.repo.UpdateTheme(ctx, id, primaryColor, secondaryColor, logoURL)
}

// UpdateCatalogFilter sets the catalog visibility filter for a tenant.
func (s *Service) UpdateCatalogFilter(ctx context.Context, id string, filter map[string]any) error {
	return s.repo.UpdateCatalogFilter(ctx, id, filter)
}

// DeleteTenant permanently removes a tenant.
func (s *Service) DeleteTenant(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ListTenants returns a paginated list of all tenants.
func (s *Service) ListTenants(ctx context.Context, limit, offset int) ([]*models.Tenant, error) {
	return s.repo.List(ctx, limit, offset)
}

func generateID() string {
	return fmt.Sprintf("tnt_%d", time.Now().UnixNano())
}
