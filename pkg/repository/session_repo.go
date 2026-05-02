package repository

import (
	"context"
	"github.com/HelixDevelopment/HelixPlay/pkg/models"
)

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByID(ctx context.Context, id string) (*models.Session, error)
	Update(ctx context.Context, session *models.Session) error
	UpdateStatus(ctx context.Context, id string, status models.SessionStatus) error
	UpdateLatency(ctx context.Context, id string, p50, p99, p999 float64) error
	Terminate(ctx context.Context, id string, reason string) error
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Session, error)
	ListByHost(ctx context.Context, hostID string, limit, offset int) ([]*models.Session, error)
	ListActive(ctx context.Context, limit, offset int) ([]*models.Session, error)
}
