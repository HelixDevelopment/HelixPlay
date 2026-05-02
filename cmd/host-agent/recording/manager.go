// Package recording manages session recording lifecycle on the host agent.
package recording

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// Status represents the recording state.
type Status string

const (
	StatusIdle     Status = "idle"
	StatusStarting Status = "starting"
	StatusRecording Status = "recording"
	StatusPaused   Status = "paused"
	StatusStopping Status = "stopping"
	StatusComplete Status = "complete"
	StatusError    Status = "error"
)

// Recording holds metadata for an active or completed recording.
type Recording struct {
	ID          string
	SessionID   string
	Status      Status
	Format      string // "mkv" or "fmp4"
	Path        string
	StartedAt   time.Time
	StoppedAt   *time.Time
	DurationMs  int64
	SizeBytes   int64
	Segments    []Segment
	mu          sync.RWMutex
}

// Segment represents a single continuous recording segment.
type Segment struct {
	Index      int
	Path       string
	StartTime  time.Time
	EndTime    time.Time
	SizeBytes  int64
}

// Manager orchestrates recording start, stop, and segment rotation.
type Manager struct {
	mu          sync.RWMutex
	recordings  map[string]*Recording
	outputDir   string
	maxSegments int
	maxDuration time.Duration
}

// Config holds recording manager parameters.
type Config struct {
	OutputDir   string
	MaxSegments int
	MaxDuration time.Duration
}

// NewManager creates a recording manager.
func NewManager(cfg Config) *Manager {
	if cfg.MaxSegments <= 0 {
		cfg.MaxSegments = 10
	}
	if cfg.MaxDuration <= 0 {
		cfg.MaxDuration = 1 * time.Hour
	}
	return &Manager{
		recordings:  make(map[string]*Recording),
		outputDir:   cfg.OutputDir,
		maxSegments: cfg.MaxSegments,
		maxDuration: cfg.MaxDuration,
	}
}

// Start begins a new recording for the given session.
func (m *Manager) Start(sessionID, format string) (*Recording, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec := &Recording{
		ID:        generateID(),
		SessionID: sessionID,
		Status:    StatusRecording,
		Format:    format,
		Path:      filepath.Join(m.outputDir, fmt.Sprintf("%s_%s.%s", sessionID, generateID(), format)),
		StartedAt: time.Now().UTC(),
	}
	m.recordings[rec.ID] = rec
	return rec, nil
}

// Stop halts a recording and finalizes metadata.
func (m *Manager) Stop(recordingID string) (*Recording, error) {
	m.mu.RLock()
	rec, ok := m.recordings[recordingID]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("recording %s not found", recordingID)
	}

	rec.mu.Lock()
	now := time.Now().UTC()
	rec.StoppedAt = &now
	rec.DurationMs = now.Sub(rec.StartedAt).Milliseconds()
	rec.Status = StatusComplete
	rec.mu.Unlock()

	return rec, nil
}

// Get retrieves a recording by ID.
func (m *Manager) Get(recordingID string) (*Recording, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rec, ok := m.recordings[recordingID]
	return rec, ok
}

// List returns all recordings.
func (m *Manager) List() []*Recording {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Recording, 0, len(m.recordings))
	for _, r := range m.recordings {
		out = append(out, r)
	}
	return out
}

// ActiveCount returns the number of currently recording sessions.
func (m *Manager) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, r := range m.recordings {
		r.mu.RLock()
		status := r.Status
		r.mu.RUnlock()
		if status == StatusRecording {
			count++
		}
	}
	return count
}

// RotateSegment starts a new segment for the given recording.
func (m *Manager) RotateSegment(recordingID string) (*Segment, error) {
	m.mu.RLock()
	rec, ok := m.recordings[recordingID]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("recording %s not found", recordingID)
	}

	rec.mu.Lock()
	defer rec.mu.Unlock()

	if rec.Status != StatusRecording {
		return nil, fmt.Errorf("recording is not active")
	}
	if len(rec.Segments) >= m.maxSegments {
		return nil, fmt.Errorf("maximum segments reached")
	}

	seg := Segment{
		Index:     len(rec.Segments),
		Path:      filepath.Join(m.outputDir, fmt.Sprintf("%s_seg%d.%s", rec.ID, len(rec.Segments), rec.Format)),
		StartTime: time.Now().UTC(),
	}
	rec.Segments = append(rec.Segments, seg)
	return &seg, nil
}

func generateID() string {
	return fmt.Sprintf("rec_%d", time.Now().UnixNano())
}
