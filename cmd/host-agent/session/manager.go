// Package session manages streaming session lifecycle on the host agent.
package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
	hostcap "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
	corecap "github.com/HelixDevelopment/HelixPlay/pkg/core/capability"
)

// Status represents the state of a streaming session.
type Status string

const (
	StatusIdle        Status = "idle"
	StatusConnecting  Status = "connecting"
	StatusNegotiating Status = "negotiating"
	StatusStreaming   Status = "streaming"
	StatusPaused      Status = "paused"
	StatusTerminated  Status = "terminated"
	StatusError       Status = "error"
)

// Session holds the runtime state of an active streaming session.
type Session struct {
	ID            string
	ClientID      string
	Status        Status
	Codec         string
	Resolution    string
	FPS           int
	TransportType string
	StartedAt     time.Time
	EndedAt       *time.Time
	mu            sync.RWMutex
}

// Manager creates, negotiates, monitors, and terminates streaming sessions.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	enc      *encoder.DualPath
	trans    transport.Transport
}

// NewManager creates a session manager with the given encoder and transport.
func NewManager(enc *encoder.DualPath, trans transport.Transport) *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		enc:      enc,
		trans:    trans,
	}
}

// Create initializes a new session in the connecting state.
func (m *Manager) Create(clientID string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess := &Session{
		ID:        generateID(),
		ClientID:  clientID,
		Status:    StatusConnecting,
		StartedAt: time.Now().UTC(),
	}
	m.sessions[sess.ID] = sess
	return sess, nil
}

// Negotiate runs capability negotiation and transitions the session to streaming.
func (m *Manager) Negotiate(sessionID string, clientCaps corecap.Capabilities) error {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session %s not found", sessionID)
	}

	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.Status != StatusConnecting {
		return fmt.Errorf("session %s is not in connecting state", sessionID)
	}

	// Use host capabilities as server caps
	hostMeta, err := hostcap.Advertise()
	if err != nil {
		return fmt.Errorf("failed to advertise host capabilities: %w", err)
	}

	serverCaps := corecap.Capabilities{
		Codecs:      hostMeta.CodecsSupported,
		Resolutions: []string{"720p", "1080p", "1440p", hostMeta.MaxResolution},
		MaxFPS:      hostMeta.RefreshRate,
	}

	negotiator := corecap.NewNegotiator()
	result := negotiator.Negotiate(clientCaps, serverCaps)
	if !result.Success {
		sess.Status = StatusError
		return fmt.Errorf("negotiation failed: %s", result.Reason)
	}

	sess.Status = StatusStreaming
	sess.Codec = result.AgreedCodecs[0]
	sess.Resolution = result.AgreedResolution
	sess.FPS = result.AgreedFPS
	return nil
}

// Get retrieves a session by ID.
func (m *Manager) Get(sessionID string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sess, ok := m.sessions[sessionID]
	return sess, ok
}

// Terminate ends a session and cleans up resources.
func (m *Manager) Terminate(sessionID string) error {
	m.mu.Lock()
	sess, ok := m.sessions[sessionID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("session %s not found", sessionID)
	}
	delete(m.sessions, sessionID)
	m.mu.Unlock()

	sess.mu.Lock()
	defer sess.mu.Unlock()
	now := time.Now().UTC()
	sess.EndedAt = &now
	sess.Status = StatusTerminated
	return nil
}

// List returns all active sessions.
func (m *Manager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, s)
	}
	return out
}

// Count returns the number of active sessions.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// Monitor checks all sessions for stale activity and terminates idle ones.
func (m *Manager) Monitor(ctx context.Context, idleThreshold time.Duration) (terminated int, err error) {
	m.mu.RLock()
	active := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		s.mu.RLock()
		status := s.Status
		s.mu.RUnlock()
		if status == StatusStreaming || status == StatusPaused {
			active = append(active, s)
		}
	}
	m.mu.RUnlock()

	now := time.Now().UTC()
	for _, s := range active {
		s.mu.RLock()
		started := s.StartedAt
		s.mu.RUnlock()
		if now.Sub(started) > idleThreshold {
			if termErr := m.Terminate(s.ID); termErr == nil {
				terminated++
			}
		}
	}
	return terminated, nil
}

func generateID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}

// Status returns the current session status safely.
func (s *Session) StatusValue() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status
}

// IsActive reports whether the session is in a non-terminal state.
func (s *Session) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status != StatusTerminated && s.Status != StatusError
}
