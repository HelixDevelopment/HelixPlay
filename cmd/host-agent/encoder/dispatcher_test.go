package encoder

import (
	"runtime"
	"testing"
)

func TestDispatcherSelectAuto(t *testing.T) {
	d := NewDispatcher("auto")
	enc, err := d.Select()
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if enc == nil {
		t.Fatal("expected non-nil encoder")
	}
}

func TestDispatcherSelectExplicit(t *testing.T) {
	d := NewDispatcher("software")
	enc, err := d.Select()
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if enc == nil {
		t.Fatal("expected non-nil encoder")
	}
	if !enc.IsStreaming() {
		// Before Start, streaming should be false
		t.Log("encoder not yet streaming (expected)")
	}
}

func TestDispatcherAvailableBackends(t *testing.T) {
	d := NewDispatcher("")
	backends := d.AvailableBackends()
	if len(backends) == 0 {
		t.Fatal("expected at least one available backend")
	}

	// Verify software is always available
	foundSoftware := false
	for _, b := range backends {
		if b == "software" {
			foundSoftware = true
		}
	}
	if !foundSoftware {
		t.Fatal("expected software encoder to always be available")
	}
}

func TestDispatcherSelectUnknown(t *testing.T) {
	d := NewDispatcher("nonexistent")
	_, err := d.Select()
	if err == nil {
		t.Fatal("expected error for unknown encoder backend")
	}
}

func TestDispatcherDetectBestPlatform(t *testing.T) {
	d := NewDispatcher("")
	best := d.detectBest()

	if runtime.GOOS == "darwin" {
		if best != "videotoolbox" {
			t.Fatalf("expected videotoolbox on darwin, got %s", best)
		}
	} else {
		// On other platforms, detectBest falls back to software
		if best != "software" {
			t.Fatalf("expected software on %s, got %s", runtime.GOOS, best)
		}
	}
}
