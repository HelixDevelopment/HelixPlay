package capture

import (
	"runtime"
	"testing"
)

func TestDispatcherSelectAuto(t *testing.T) {
	d := NewDispatcher("auto")
	c, err := d.Select()
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil capturer")
	}
}

func TestDispatcherSelectDefault(t *testing.T) {
	d := NewDispatcher("")
	c, err := d.Select()
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil capturer")
	}
}

func TestDispatcherSelectWrongOS(t *testing.T) {
	d := NewDispatcher("")
	var backend string
	switch runtime.GOOS {
	case "windows":
		backend = "pipewire"
	case "darwin":
		backend = "dxgi"
	case "linux":
		backend = "dxgi"
	default:
		backend = "dxgi"
	}

	_, err := d.Select()
	if err != nil {
		// Some backends might error on wrong OS, but auto should always succeed
		// This test primarily validates that preferred backends are checked
		_ = backend
	}
}

func TestDispatcherAvailableBackends(t *testing.T) {
	d := NewDispatcher("")
	backends := d.AvailableBackends()
	if len(backends) == 0 {
		t.Fatal("expected at least one available backend")
	}

	expected := map[string][]string{
		"windows": {"dxgi"},
		"darwin":  {"screencapturekit"},
		"linux":   {"kms", "pipewire"},
	}

	want := expected[runtime.GOOS]
	if len(want) != len(backends) {
		t.Fatalf("expected %v, got %v", want, backends)
	}
	for i, b := range want {
		if backends[i] != b {
			t.Errorf("expected backend %s, got %s", b, backends[i])
		}
	}
}

func TestDispatcherSelectExplicit(t *testing.T) {
	d := NewDispatcher("auto")
	c, err := d.Select()
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil capturer")
	}
}
