package repository

import (
	"context"
	"github.com/HelixDevelopment/HelixPlay/pkg/models"
)

// GameRepository defines the interface for game data access
type GameRepository interface {
	Create(ctx context.Context, game *models.Game) error
	GetByID(ctx context.Context, id string) (*models.Game, error)
	GetByStoreID(ctx context.Context, storeType models.GameStoreType, storeID string) (*models.Game, error)
	Update(ctx context.Context, game *models.Game) error
	Delete(ctx context.Context, id string) error
	ListByHost(ctx context.Context, hostID string, limit, offset int) ([]*models.Game, error)
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Game, error)
	Search(ctx context.Context, tenantID, query string, limit, offset int) ([]*models.Game, error)
}
