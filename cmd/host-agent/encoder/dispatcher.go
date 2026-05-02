// Package encoder provides hardware-accelerated encoding with automatic backend selection.
package encoder

import (
	"fmt"
	"runtime"
	"strings"
)

// Dispatcher selects the best encoder backend for the detected GPU.
type Dispatcher struct {
	preferred string
}

// NewDispatcher creates an encoder dispatcher with optional preferred backend.
func NewDispatcher(preferred string) *Dispatcher {
	return &Dispatcher{preferred: preferred}
}

// Select returns the best encoder for the current hardware.
func (d *Dispatcher) Select() (*DualPath, error) {
	backend := d.preferred
	if backend == "" || backend == "auto" {
		backend = d.detectBest()
	}

	switch strings.ToLower(backend) {
	case "nvenc", "qsv", "amf", "videotoolbox", "vaapi", "vulkan", "software", "sw":
		enc := NewDualPath(backend)
		if enc == nil {
			return nil, fmt.Errorf("encoder %s not available", backend)
		}
		return enc, nil
	default:
		enc := NewDualPath(backend)
		if enc == nil {
			return nil, fmt.Errorf("unknown encoder backend: %s", backend)
		}
		return enc, nil
	}
}

// AvailableBackends returns encoder backends supported on this platform.
func (d *Dispatcher) AvailableBackends() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"nvenc", "amf", "qsv", "software"}
	case "darwin":
		return []string{"videotoolbox", "software"}
	case "linux":
		return []string{"nvenc", "vaapi", "qsv", "vulkan", "software"}
	default:
		return []string{"software"}
	}
}

// detectBest returns the best encoder based on simple heuristics.
func (d *Dispatcher) detectBest() string {
	// Apple Silicon
	if runtime.GOOS == "darwin" {
		return "videotoolbox"
	}
	// GPU detection would check sysfs/registry; in test environments
	// we fall back to software encoding.
	return "software"
}
