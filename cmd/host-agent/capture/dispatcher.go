// Package capture provides OS-specific screen capture backends and dispatch.
package capture

import (
	"fmt"
	"runtime"
)

// Dispatcher selects the best capture backend for the current platform.
type Dispatcher struct {
	preferred string
}

// NewDispatcher creates a capture dispatcher with optional preferred backend.
// If preferred is empty or "auto", the best backend for the current OS is selected.
func NewDispatcher(preferred string) *Dispatcher {
	return &Dispatcher{preferred: preferred}
}

// Select returns the best capturer for the current OS.
func (d *Dispatcher) Select() (Capturer, error) {
	backend := d.preferred
	if backend == "" || backend == "auto" {
		backend = defaultBackend()
	}

	switch backend {
	case "dxgi":
		if runtime.GOOS != "windows" {
			return nil, fmt.Errorf("DXGI backend requires Windows")
		}
		return newPlatformCapturer(), nil
	case "screencapturekit", "sck":
		if runtime.GOOS != "darwin" {
			return nil, fmt.Errorf("ScreenCaptureKit backend requires macOS")
		}
		return newPlatformCapturer(), nil
	case "pipewire":
		if runtime.GOOS != "linux" {
			return nil, fmt.Errorf("PipeWire backend requires Linux")
		}
		return newPlatformCapturer(), nil
	case "kms":
		if runtime.GOOS != "linux" {
			return nil, fmt.Errorf("KMS backend requires Linux")
		}
		// KMS/DMA-BUF is Linux-only; fall through to platform capturer
		return newPlatformCapturer(), nil
	default:
		return NewCapturer(runtime.GOOS), nil
	}
}

// AvailableBackends returns the list of capture backends supported on this platform.
func (d *Dispatcher) AvailableBackends() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"dxgi"}
	case "darwin":
		return []string{"screencapturekit"}
	case "linux":
		return []string{"kms", "pipewire"}
	default:
		return []string{}
	}
}

func defaultBackend() string {
	switch runtime.GOOS {
	case "windows":
		return "dxgi"
	case "darwin":
		return "screencapturekit"
	case "linux":
		return "pipewire"
	default:
		return ""
	}
}
