//go:build linux

package capture

import (
	"fmt"
	"os"
)

type pipeWireCapturer struct {
	baseCapturer
}

func newPlatformCapturer() Capturer {
	return &pipeWireCapturer{
		baseCapturer: baseCapturer{osType: "linux"},
	}
}

func (p *pipeWireCapturer) Start() error {
	if !isPipeWireAvailable() {
		return fmt.Errorf("pipewire capture unavailable: pipewire socket not found (install pipewire and ensure /run/user/*/pipewire-0 exists)")
	}
	p.running = true
	return nil
}

func (p *pipeWireCapturer) Stop() {
	p.running = false
}

func (p *pipeWireCapturer) IsRunning() bool {
	return p.running
}

func (p *pipeWireCapturer) GetFrame() ([]byte, error) {
	if !p.running {
		return nil, fmt.Errorf("capturer not running")
	}
	return nil, fmt.Errorf("pipewire capture requires CGo bindings to libpipewire-0.3 (not available in this build)")
}

func isPipeWireAvailable() bool {
	pipewirePaths := []string{
		"/run/user/1000/pipewire-0",
		"/run/user/0/pipewire-0",
	}
	for _, path := range pipewirePaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}
