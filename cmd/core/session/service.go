package session

import (
	"context"
	"fmt"
	"time"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/capability"
	"github.com/HelixDevelopment/HelixPlay/pkg/models"
	"github.com/HelixDevelopment/HelixPlay/pkg/repository"
)

// Service handles the lifecycle of streaming sessions.
type Service struct {
	sessions   repository.SessionRepository
	hosts      repository.HostRepository
	negotiator *capability.Negotiator
	discovery  *capability.Discovery
}

// NewService creates a session service backed by the given repositories.
func NewService(
	sessions repository.SessionRepository,
	hosts repository.HostRepository,
	discovery *capability.Discovery,
) *Service {
	return &Service{
		sessions:   sessions,
		hosts:      hosts,
		negotiator: capability.NewNegotiator(),
		discovery:  discovery,
	}
}

// CreateSession initializes a new streaming session for a user and host.
func (s *Service) CreateSession(ctx context.Context, userID, hostID string, gameID *string, controllerType string, recordingEnabled bool) (*models.Session, error) {
	now := time.Now().UTC()
	session := &models.Session{
		ID:               generateID(),
		UserID:           userID,
		HostID:           hostID,
		GameID:           gameID,
		Status:           models.SessionStatusConnecting,
		ControllerType:   controllerType,
		RecordingEnabled: recordingEnabled,
		StartedAt:        now,
		LastActivityAt:   now,
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// NegotiateSession runs capability negotiation between client and server and
// updates the session with the agreed parameters.
func (s *Service) NegotiateSession(ctx context.Context, sessionID string, client, server capability.Capabilities) (*models.Session, error) {
	session, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != models.SessionStatusConnecting {
		return nil, fmt.Errorf("session %s is not in connecting state", sessionID)
	}

	result := s.negotiator.Negotiate(client, server)
	if !result.Success {
		_ = s.sessions.UpdateStatus(ctx, sessionID, models.SessionStatusEnded)
		return nil, fmt.Errorf("negotiation failed: %s", result.Reason)
	}

	session.Status = models.SessionStatusNegotiating
	session.Codec = result.AgreedCodecs[0]
	session.Resolution = result.AgreedResolution
	session.RefreshHz = result.AgreedFPS
	session.LastActivityAt = time.Now().UTC()

	if err := s.sessions.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session during negotiation: %w", err)
	}

	if err := s.sessions.UpdateStatus(ctx, sessionID, models.SessionStatusStreaming); err != nil {
		return nil, fmt.Errorf("failed to transition to streaming: %w", err)
	}

	session.Status = models.SessionStatusStreaming
	return session, nil
}

// GetSession retrieves a session by its ID.
func (s *Service) GetSession(ctx context.Context, id string) (*models.Session, error) {
	return s.sessions.GetByID(ctx, id)
}

// UpdateLatency records latency percentiles for an active session.
func (s *Service) UpdateLatency(ctx context.Context, id string, p50, p99, p999 float64) error {
	return s.sessions.UpdateLatency(ctx, id, p50, p99, p999)
}

// TerminateSession ends a session and records the termination reason.
func (s *Service) TerminateSession(ctx context.Context, id, reason string) error {
	return s.sessions.Terminate(ctx, id, reason)
}

// ListByUser returns sessions for a specific user.
func (s *Service) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Session, error) {
	return s.sessions.ListByUser(ctx, userID, limit, offset)
}

// ListByHost returns sessions for a specific host.
func (s *Service) ListByHost(ctx context.Context, hostID string, limit, offset int) ([]*models.Session, error) {
	return s.sessions.ListByHost(ctx, hostID, limit, offset)
}

// ListActive returns currently active sessions.
func (s *Service) ListActive(ctx context.Context, limit, offset int) ([]*models.Session, error) {
	return s.sessions.ListActive(ctx, limit, offset)
}

// MonitorSessions checks active sessions for stale activity and terminates
// any that have exceeded the given idle threshold.
func (s *Service) MonitorSessions(ctx context.Context, idleThreshold time.Duration) (terminated int, err error) {
	sessions, err := s.sessions.ListActive(ctx, 1000, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to list active sessions: %w", err)
	}

	now := time.Now().UTC()
	for _, sess := range sessions {
		if now.Sub(sess.LastActivityAt) > idleThreshold {
			if termErr := s.sessions.Terminate(ctx, sess.ID, "idle timeout"); termErr == nil {
				terminated++
			}
		}
	}
	return terminated, nil
}

// DiscoverHostAgents returns available host-agent endpoints from discovery.
func (s *Service) DiscoverHostAgents() ([]capability.Endpoint, error) {
	return s.discovery.Discover()
}

func generateID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}
