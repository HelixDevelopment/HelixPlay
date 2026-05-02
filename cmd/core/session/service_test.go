package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/session"
	"github.com/HelixDevelopment/HelixPlay/pkg/models"
)

// mockSessionRepository is an in-memory implementation of repository.SessionRepository.
type mockSessionRepository struct {
	sessions map[string]*models.Session
}

func newMockSessionRepository() *mockSessionRepository {
	return &mockSessionRepository{sessions: make(map[string]*models.Session)}
}

func (m *mockSessionRepository) Create(ctx context.Context, s *models.Session) error {
	m.sessions[s.ID] = s
	return nil
}

func (m *mockSessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return s, nil
}

func (m *mockSessionRepository) Update(ctx context.Context, s *models.Session) error {
	m.sessions[s.ID] = s
	return nil
}

func (m *mockSessionRepository) UpdateStatus(ctx context.Context, id string, status models.SessionStatus) error {
	s, ok := m.sessions[id]
	if !ok {
		return errors.New("not found")
	}
	s.Status = status
	return nil
}

func (m *mockSessionRepository) UpdateLatency(ctx context.Context, id string, p50, p99, p999 float64) error {
	s, ok := m.sessions[id]
	if !ok {
		return errors.New("not found")
	}
	s.LatencyP50Ms = &p50
	s.LatencyP99Ms = &p99
	s.LatencyP999Ms = &p999
	return nil
}

func (m *mockSessionRepository) Terminate(ctx context.Context, id string, reason string) error {
	s, ok := m.sessions[id]
	if !ok {
		return errors.New("not found")
	}
	s.Status = models.SessionStatusEnded
	now := time.Now().UTC()
	s.EndedAt = &now
	return nil
}

func (m *mockSessionRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Session, error) {
	var out []*models.Session
	for _, s := range m.sessions {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (m *mockSessionRepository) ListByHost(ctx context.Context, hostID string, limit, offset int) ([]*models.Session, error) {
	var out []*models.Session
	for _, s := range m.sessions {
		if s.HostID == hostID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (m *mockSessionRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.Session, error) {
	var out []*models.Session
	for _, s := range m.sessions {
		if s.Status == models.SessionStatusStreaming || s.Status == models.SessionStatusNegotiating || s.Status == models.SessionStatusConnecting {
			out = append(out, s)
		}
	}
	return out, nil
}

// mockHostRepository is a minimal stub satisfying repository.HostRepository.
type mockHostRepository struct{}

func (m *mockHostRepository) Create(ctx context.Context, host *models.Host) error         { return nil }
func (m *mockHostRepository) GetByID(ctx context.Context, id string) (*models.Host, error) { return nil, errors.New("not found") }
func (m *mockHostRepository) GetByHardwareID(ctx context.Context, hardwareID string) (*models.Host, error) {
	return nil, errors.New("not found")
}
func (m *mockHostRepository) Update(ctx context.Context, host *models.Host) error                { return nil }
func (m *mockHostRepository) UpdateStatus(ctx context.Context, id string, status models.HostStatus) error { return nil }
func (m *mockHostRepository) Delete(ctx context.Context, id string) error                        { return nil }
func (m *mockHostRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Host, error) {
	return nil, nil
}
func (m *mockHostRepository) ListOnline(ctx context.Context, limit, offset int) ([]*models.Host, error) { return nil, nil }

func TestCreateSession(t *testing.T) {
	ctx := context.Background()
	repo := newMockSessionRepository()
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	svc := session.NewService(repo, &mockHostRepository{}, disc)

	gameID := "game-1"
	sess, err := svc.CreateSession(ctx, "user-1", "host-1", &gameID, "dualsense", true)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if sess.UserID != "user-1" {
		t.Errorf("expected userID user-1, got %s", sess.UserID)
	}
	if sess.Status != models.SessionStatusConnecting {
		t.Errorf("expected status connecting, got %s", sess.Status)
	}
	if sess.GameID == nil || *sess.GameID != "game-1" {
		t.Error("expected GameID to be set")
	}

	stored, err := repo.GetByID(ctx, sess.ID)
	if err != nil {
		t.Fatalf("session not stored: %v", err)
	}
	if stored.ID != sess.ID {
		t.Error("stored session ID mismatch")
	}
}

func TestNegotiateSession_Success(t *testing.T) {
	ctx := context.Background()
	repo := newMockSessionRepository()
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	svc := session.NewService(repo, &mockHostRepository{}, disc)

	sess, _ := svc.CreateSession(ctx, "user-1", "host-1", nil, "xbox", false)

	clientCaps := capability.Capabilities{
		Codecs:      []string{"H.264", "HEVC"},
		Resolutions: []string{"1080p", "4K"},
		MaxFPS:      120,
	}
	serverCaps := capability.Capabilities{
		Codecs:      []string{"H.264", "HEVC", "AV1"},
		Resolutions: []string{"720p", "1080p", "1440p", "4K"},
		MaxFPS:      240,
	}

	result, err := svc.NegotiateSession(ctx, sess.ID, clientCaps, serverCaps)
	if err != nil {
		t.Fatalf("NegotiateSession failed: %v", err)
	}
	if result.Status != models.SessionStatusStreaming {
		t.Errorf("expected streaming status, got %s", result.Status)
	}
	if result.Codec != "H.264" {
		t.Errorf("expected codec H.264, got %s", result.Codec)
	}
	if result.Resolution != "4K" {
		t.Errorf("expected resolution 4K, got %s", result.Resolution)
	}
	if result.RefreshHz != 120 {
		t.Errorf("expected refresh 120, got %d", result.RefreshHz)
	}
}

func TestNegotiateSession_FailureNoCommonCodec(t *testing.T) {
	ctx := context.Background()
	repo := newMockSessionRepository()
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	svc := session.NewService(repo, &mockHostRepository{}, disc)

	sess, _ := svc.CreateSession(ctx, "user-1", "host-1", nil, "xbox", false)

	clientCaps := capability.Capabilities{Codecs: []string{"VP8"}}
	serverCaps := capability.Capabilities{Codecs: []string{"H.264", "HEVC"}}

	_, err := svc.NegotiateSession(ctx, sess.ID, clientCaps, serverCaps)
	if err == nil {
		t.Fatal("expected negotiation failure")
	}

	stored, _ := repo.GetByID(ctx, sess.ID)
	if stored.Status != models.SessionStatusEnded {
		t.Errorf("expected ended status after failed negotiation, got %s", stored.Status)
	}
}

func TestNegotiateSession_WrongState(t *testing.T) {
	ctx := context.Background()
	repo := newMockSessionRepository()
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	svc := session.NewService(repo, &mockHostRepository{}, disc)

	sess, _ := svc.CreateSession(ctx, "user-1", "host-1", nil, "xbox", false)
	_ = repo.UpdateStatus(ctx, sess.ID, models.SessionStatusStreaming)

	_, err := svc.NegotiateSession(ctx, sess.ID, capability.Capabilities{}, capability.Capabilities{})
	if err == nil {
		t.Fatal("expected error when negotiating non-connecting session")
	}
}

func TestUpdateLatency(t *testing.T) {
	ctx := context.Background()
	repo := newMockSessionRepository()
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	svc := session.NewService(repo, &mockHostRepository{}, disc)

	sess, _ := svc.CreateSession(ctx, "user-1", "host-1", nil, "xbox", false)

	if err := svc.UpdateLatency(ctx, sess.ID, 12.5, 45.0, 89.0); err != nil {
		t.Fatalf("UpdateLatency failed: %v", err)
	}

	stored, _ := repo.GetByID(ctx, sess.ID)
	if stored.LatencyP50Ms == nil || *stored.LatencyP50Ms != 12.5 {
		t.Errorf("expected p50 12.5, got %v", stored.LatencyP50Ms)
	}
	if stored.LatencyP99Ms == nil || *stored.LatencyP99Ms != 45.0 {
		t.Errorf("expected p99 45.0, got %v", stored.LatencyP99Ms)
	}
	if stored.LatencyP999Ms == nil || *stored.LatencyP999Ms != 89.0 {
		t.Errorf("expected p999 89.0, got %v", stored.LatencyP999Ms)
	}
}

func TestTerminateSession(t *testing.T) {
	ctx := context.Background()
	repo := newMockSessionRepository()
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	svc := session.NewService(repo, &mockHostRepository{}, disc)

	sess, _ := svc.CreateSession(ctx, "user-1", "host-1", nil, "xbox", false)

	if err := svc.TerminateSession(ctx, sess.ID, "user logout"); err != nil {
		t.Fatalf("TerminateSession failed: %v", err)
	}

	stored, _ := repo.GetByID(ctx, sess.ID)
	if stored.Status != models.SessionStatusEnded {
		t.Errorf("expected ended status, got %s", stored.Status)
	}
	if stored.EndedAt == nil {
		t.Error("expected EndedAt to be set")
	}
}

func TestListByUser(t *testing.T) {
	ctx := context.Background()
	repo := newMockSessionRepository()
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	svc := session.NewService(repo, &mockHostRepository{}, disc)

	svc.CreateSession(ctx, "alice", "host-1", nil, "xbox", false)
	svc.CreateSession(ctx, "alice", "host-2", nil, "ps5", false)
	svc.CreateSession(ctx, "bob", "host-1", nil, "xbox", false)

	sessions, err := svc.ListByUser(ctx, "alice", 10, 0)
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions for alice, got %d", len(sessions))
	}
}

func TestListActive(t *testing.T) {
	ctx := context.Background()
	repo := newMockSessionRepository()
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	svc := session.NewService(repo, &mockHostRepository{}, disc)

	s1, _ := svc.CreateSession(ctx, "user-1", "host-1", nil, "xbox", false)
	s2, _ := svc.CreateSession(ctx, "user-2", "host-1", nil, "ps5", false)
	s3, _ := svc.CreateSession(ctx, "user-3", "host-1", nil, "xbox", false)
	_ = repo.UpdateStatus(ctx, s1.ID, models.SessionStatusStreaming)
	_ = repo.UpdateStatus(ctx, s2.ID, models.SessionStatusStreaming)
	_ = repo.UpdateStatus(ctx, s3.ID, models.SessionStatusEnded)

	active, err := svc.ListActive(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListActive failed: %v", err)
	}
	if len(active) != 2 {
		t.Errorf("expected 2 active sessions, got %d", len(active))
	}
}

func TestMonitorSessions(t *testing.T) {
	ctx := context.Background()
	repo := newMockSessionRepository()
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	svc := session.NewService(repo, &mockHostRepository{}, disc)

	s1, _ := svc.CreateSession(ctx, "user-1", "host-1", nil, "xbox", false)
	_ = repo.UpdateStatus(ctx, s1.ID, models.SessionStatusStreaming)
	// artificially age last activity
	repo.sessions[s1.ID].LastActivityAt = time.Now().UTC().Add(-10 * time.Minute)

	s2, _ := svc.CreateSession(ctx, "user-2", "host-1", nil, "ps5", false)
	_ = repo.UpdateStatus(ctx, s2.ID, models.SessionStatusStreaming)
	repo.sessions[s2.ID].LastActivityAt = time.Now().UTC().Add(-1 * time.Minute)

	terminated, err := svc.MonitorSessions(ctx, 5*time.Minute)
	if err != nil {
		t.Fatalf("MonitorSessions failed: %v", err)
	}
	if terminated != 1 {
		t.Errorf("expected 1 termination, got %d", terminated)
	}

	stored, _ := repo.GetByID(ctx, s1.ID)
	if stored.Status != models.SessionStatusEnded {
		t.Errorf("expected s1 ended, got %s", stored.Status)
	}
}

func TestDiscoverHostAgents(t *testing.T) {
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	svc := session.NewService(nil, nil, disc)

	endpoints, err := svc.DiscoverHostAgents()
	if err != nil {
		t.Fatalf("DiscoverHostAgents failed: %v", err)
	}
	if len(endpoints) == 0 {
		t.Error("expected at least one endpoint")
	}
	if endpoints[0].Type != "host-agent" {
		t.Errorf("expected type host-agent, got %s", endpoints[0].Type)
	}
}
