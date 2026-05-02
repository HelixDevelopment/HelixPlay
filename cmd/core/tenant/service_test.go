package tenant_test

import (
	"context"
	"errors"
	"testing"

	"github.com/HelixDevelopment/HelixPlay/cmd/core/tenant"
	"github.com/HelixDevelopment/HelixPlay/pkg/models"
)

// mockTenantRepository is an in-memory implementation of repository.TenantRepository.
type mockTenantRepository struct {
	tenants map[string]*models.Tenant
}

func newMockTenantRepository() *mockTenantRepository {
	return &mockTenantRepository{tenants: make(map[string]*models.Tenant)}
}

func (m *mockTenantRepository) Create(ctx context.Context, t *models.Tenant) error {
	m.tenants[t.ID] = t
	return nil
}

func (m *mockTenantRepository) GetByID(ctx context.Context, id string) (*models.Tenant, error) {
	t, ok := m.tenants[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}

func (m *mockTenantRepository) GetBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	for _, t := range m.tenants {
		if t.Slug == slug {
			return t, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockTenantRepository) Update(ctx context.Context, t *models.Tenant) error {
	m.tenants[t.ID] = t
	return nil
}

func (m *mockTenantRepository) UpdateTheme(ctx context.Context, id, primaryColor, secondaryColor, logoURL string) error {
	t, ok := m.tenants[id]
	if !ok {
		return errors.New("not found")
	}
	t.ThemePrimaryColor = primaryColor
	t.ThemeSecondaryColor = secondaryColor
	t.ThemeLogoURL = logoURL
	return nil
}

func (m *mockTenantRepository) UpdateCatalogFilter(ctx context.Context, id string, filter map[string]any) error {
	t, ok := m.tenants[id]
	if !ok {
		return errors.New("not found")
	}
	t.CatalogFilter = filter
	return nil
}

func (m *mockTenantRepository) Delete(ctx context.Context, id string) error {
	delete(m.tenants, id)
	return nil
}

func (m *mockTenantRepository) List(ctx context.Context, limit, offset int) ([]*models.Tenant, error) {
	var out []*models.Tenant
	for _, t := range m.tenants {
		out = append(out, t)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func TestCreateTenant_Validation(t *testing.T) {
	ctx := context.Background()
	repo := newMockTenantRepository()
	svc := tenant.NewService(repo)

	_, err := svc.CreateTenant(ctx, &models.Tenant{Slug: "", Name: "Test"})
	if err == nil {
		t.Error("expected error for empty slug")
	}

	_, err = svc.CreateTenant(ctx, &models.Tenant{Slug: "test", Name: ""})
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestCreateTenant_Success(t *testing.T) {
	ctx := context.Background()
	repo := newMockTenantRepository()
	svc := tenant.NewService(repo)

	tn, err := svc.CreateTenant(ctx, &models.Tenant{Slug: "acme", Name: "ACME Corp"})
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}
	if tn.ID == "" {
		t.Error("expected tenant ID to be generated")
	}
	if tn.Slug != "acme" {
		t.Errorf("expected slug acme, got %s", tn.Slug)
	}

	stored, err := repo.GetByID(ctx, tn.ID)
	if err != nil {
		t.Fatalf("tenant not stored: %v", err)
	}
	if stored.Name != "ACME Corp" {
		t.Errorf("expected name ACME Corp, got %s", stored.Name)
	}
}

func TestGetTenant(t *testing.T) {
	ctx := context.Background()
	repo := newMockTenantRepository()
	svc := tenant.NewService(repo)

	tn, _ := svc.CreateTenant(ctx, &models.Tenant{Slug: "acme", Name: "ACME"})

	found, err := svc.GetTenant(ctx, tn.ID)
	if err != nil {
		t.Fatalf("GetTenant failed: %v", err)
	}
	if found.ID != tn.ID {
		t.Error("ID mismatch")
	}

	_, err = svc.GetTenant(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for missing tenant")
	}
}

func TestGetTenantBySlug(t *testing.T) {
	ctx := context.Background()
	repo := newMockTenantRepository()
	svc := tenant.NewService(repo)

	_, _ = svc.CreateTenant(ctx, &models.Tenant{Slug: "acme", Name: "ACME"})

	found, err := svc.GetTenantBySlug(ctx, "acme")
	if err != nil {
		t.Fatalf("GetTenantBySlug failed: %v", err)
	}
	if found.Slug != "acme" {
		t.Errorf("expected slug acme, got %s", found.Slug)
	}
}

func TestUpdateTenant(t *testing.T) {
	ctx := context.Background()
	repo := newMockTenantRepository()
	svc := tenant.NewService(repo)

	tn, _ := svc.CreateTenant(ctx, &models.Tenant{Slug: "acme", Name: "ACME"})

	updated, err := svc.UpdateTenant(ctx, &models.Tenant{
		ID:   tn.ID,
		Slug: "acme-new",
		Name: "ACME New",
	})
	if err != nil {
		t.Fatalf("UpdateTenant failed: %v", err)
	}
	if updated.Name != "ACME New" {
		t.Errorf("expected name ACME New, got %s", updated.Name)
	}

	stored, _ := repo.GetByID(ctx, tn.ID)
	if stored.Slug != "acme-new" {
		t.Errorf("expected slug acme-new, got %s", stored.Slug)
	}
}

func TestUpdateTheme(t *testing.T) {
	ctx := context.Background()
	repo := newMockTenantRepository()
	svc := tenant.NewService(repo)

	tn, _ := svc.CreateTenant(ctx, &models.Tenant{Slug: "acme", Name: "ACME"})

	if err := svc.UpdateTheme(ctx, tn.ID, "#ff0000", "#00ff00", "https://logo.png"); err != nil {
		t.Fatalf("UpdateTheme failed: %v", err)
	}

	stored, _ := repo.GetByID(ctx, tn.ID)
	if stored.ThemePrimaryColor != "#ff0000" {
		t.Errorf("expected primary #ff0000, got %s", stored.ThemePrimaryColor)
	}
	if stored.ThemeLogoURL != "https://logo.png" {
		t.Errorf("expected logo URL, got %s", stored.ThemeLogoURL)
	}
}

func TestUpdateCatalogFilter(t *testing.T) {
	ctx := context.Background()
	repo := newMockTenantRepository()
	svc := tenant.NewService(repo)

	tn, _ := svc.CreateTenant(ctx, &models.Tenant{Slug: "acme", Name: "ACME"})

	filter := map[string]any{"genre": "action", "hdr": true}
	if err := svc.UpdateCatalogFilter(ctx, tn.ID, filter); err != nil {
		t.Fatalf("UpdateCatalogFilter failed: %v", err)
	}

	stored, _ := repo.GetByID(ctx, tn.ID)
	if stored.CatalogFilter == nil {
		t.Fatal("expected catalog filter to be set")
	}
	if stored.CatalogFilter["genre"] != "action" {
		t.Errorf("expected genre action, got %v", stored.CatalogFilter["genre"])
	}
}

func TestDeleteTenant(t *testing.T) {
	ctx := context.Background()
	repo := newMockTenantRepository()
	svc := tenant.NewService(repo)

	tn, _ := svc.CreateTenant(ctx, &models.Tenant{Slug: "acme", Name: "ACME"})

	if err := svc.DeleteTenant(ctx, tn.ID); err != nil {
		t.Fatalf("DeleteTenant failed: %v", err)
	}

	_, err := repo.GetByID(ctx, tn.ID)
	if err == nil {
		t.Error("expected tenant to be deleted")
	}
}

func TestListTenants(t *testing.T) {
	ctx := context.Background()
	repo := newMockTenantRepository()
	svc := tenant.NewService(repo)

	svc.CreateTenant(ctx, &models.Tenant{Slug: "a", Name: "A"})
	svc.CreateTenant(ctx, &models.Tenant{Slug: "b", Name: "B"})

	list, err := svc.ListTenants(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListTenants failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 tenants, got %d", len(list))
	}
}
