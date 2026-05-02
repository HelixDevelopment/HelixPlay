package repository

import (
	"context"
	"github.com/HelixDevelopment/HelixPlay/pkg/models"
)

// HostRepository defines the interface for host data access
type HostRepository interface {
	Create(ctx context.Context, host *models.Host) error
	GetByID(ctx context.Context, id string) (*models.Host, error)
	GetByHardwareID(ctx context.Context, hardwareID string) (*models.Host, error)
	Update(ctx context.Context, host *models.Host) error
	UpdateStatus(ctx context.Context, id string, status models.HostStatus) error
	Delete(ctx context.Context, id string) error
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Host, error)
	ListOnline(ctx context.Context, limit, offset int) ([]*models.Host, error)
}
