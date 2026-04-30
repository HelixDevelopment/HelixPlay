//go:build linux

package capture

import "fmt"

type pipeWireCapturer struct {
	baseCapturer
}

func newPlatformCapturer() Capturer {
	return &pipeWireCapturer{
		baseCapturer: baseCapturer{osType: "linux"},
	}
}

func (p *pipeWireCapturer) Start() error {
	// Stub: Initialize PipeWire capture
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
	return nil, fmt.Errorf("not implemented")
}
