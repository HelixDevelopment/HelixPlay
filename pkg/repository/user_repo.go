package repository

import (
	"context"
	"github.com/HelixDevelopment/HelixPlay/pkg/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, tenantID, email string) (*models.User, error)
	GetByOAuthSubject(ctx context.Context, tenantID, provider, subject string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateLastLogin(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.User, error)
}
