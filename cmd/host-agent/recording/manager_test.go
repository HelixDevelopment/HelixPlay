package recording

import (
	"path/filepath"
	"testing"
)

func TestManagerStartStop(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(Config{OutputDir: tmpDir})

	rec, err := m.Start("sess-1", "mkv")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if rec.ID == "" {
		t.Fatal("expected non-empty recording ID")
	}
	if rec.Status != StatusRecording {
		t.Errorf("expected status recording, got %s", rec.Status)
	}
	if rec.SessionID != "sess-1" {
		t.Errorf("expected sess-1, got %s", rec.SessionID)
	}

	stopped, err := m.Stop(rec.ID)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if stopped.Status != StatusComplete {
		t.Errorf("expected status complete, got %s", stopped.Status)
	}
	if stopped.DurationMs < 0 {
		t.Error("expected non-negative duration")
	}
}

func TestManagerGetList(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(Config{OutputDir: tmpDir})

	rec1, _ := m.Start("sess-1", "mkv")
	_, _ = m.Start("sess-2", "fmp4")

	list := m.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 recordings, got %d", len(list))
	}

	got, ok := m.Get(rec1.ID)
	if !ok {
		t.Fatal("expected recording to be found")
	}
	if got.ID != rec1.ID {
		t.Errorf("expected %s, got %s", rec1.ID, got.ID)
	}

	_, ok = m.Get("nonexistent")
	if ok {
		t.Error("expected recording not to be found")
	}
}

func TestManagerActiveCount(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(Config{OutputDir: tmpDir})

	if m.ActiveCount() != 0 {
		t.Fatalf("expected 0 active, got %d", m.ActiveCount())
	}

	rec, _ := m.Start("sess-1", "mkv")
	if m.ActiveCount() != 1 {
		t.Fatalf("expected 1 active, got %d", m.ActiveCount())
	}

	m.Stop(rec.ID)
	if m.ActiveCount() != 0 {
		t.Fatalf("expected 0 active after stop, got %d", m.ActiveCount())
	}
}

func TestManagerRotateSegment(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(Config{OutputDir: tmpDir, MaxSegments: 3})

	rec, _ := m.Start("sess-1", "mkv")

	seg1, err := m.RotateSegment(rec.ID)
	if err != nil {
		t.Fatalf("RotateSegment failed: %v", err)
	}
	if seg1.Index != 0 {
		t.Errorf("expected index 0, got %d", seg1.Index)
	}

	seg2, err := m.RotateSegment(rec.ID)
	if err != nil {
		t.Fatalf("RotateSegment failed: %v", err)
	}
	if seg2.Index != 1 {
		t.Errorf("expected index 1, got %d", seg2.Index)
	}

	// MaxSegments = 3, we already have 2 segments, one more allowed
	_, err = m.RotateSegment(rec.ID)
	if err != nil {
		t.Fatalf("RotateSegment failed: %v", err)
	}

	// This should fail (4th segment)
	_, err = m.RotateSegment(rec.ID)
	if err == nil {
		t.Fatal("expected error for exceeding max segments")
	}
}

func TestManagerRotateNotActive(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(Config{OutputDir: tmpDir})

	rec, _ := m.Start("sess-1", "mkv")
	m.Stop(rec.ID)

	_, err := m.RotateSegment(rec.ID)
	if err == nil {
		t.Fatal("expected error for stopped recording")
	}
}

func TestManagerPathGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(Config{OutputDir: tmpDir})

	rec, _ := m.Start("sess-1", "mkv")
	if rec.Path == "" {
		t.Fatal("expected non-empty path")
	}
	if filepath.Dir(rec.Path) != tmpDir {
		t.Errorf("expected path in %s, got %s", tmpDir, rec.Path)
	}
	if filepath.Ext(rec.Path) != ".mkv" {
		t.Errorf("expected .mkv extension, got %s", filepath.Ext(rec.Path))
	}
}
