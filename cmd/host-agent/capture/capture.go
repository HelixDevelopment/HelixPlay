package capture

import (
	"fmt"
	"runtime"
)

// Capturer interface defines the contract for screen capture backends
type Capturer interface {
	Start() error
	Stop()
	IsRunning() bool
	GetFrame() ([]byte, error)
}

// baseCapturer provides common fields for all capturer implementations
type baseCapturer struct {
	running bool
	osType  string
}

// NewCapturer creates a new capturer for the specified OS type
// Returns an unsupportedCapturer if the OS type doesn't match the current platform
func NewCapturer(osType string) Capturer {
	if osType == runtime.GOOS {
		return newPlatformCapturer()
	}
	return &unsupportedCapturer{osType: osType}
}

// unsupportedCapturer handles unsupported OS types or mismatched OS requests
type unsupportedCapturer struct {
	osType string
}

func (u *unsupportedCapturer) Start() error {
	return fmt.Errorf("unsupported OS: %s", u.osType)
}

func (u *unsupportedCapturer) Stop() {
	// no-op: unsupported capturer has no resources to release
}
func (u *unsupportedCapturer) IsRunning() bool { return false }
func (u *unsupportedCapturer) GetFrame() ([]byte, error) {
	return nil, fmt.Errorf("unsupported OS: %s (no capture backend available)", u.osType)
}
