//go:build windows

package capture

import "fmt"

type dx11Capturer struct {
	baseCapturer
}

func newPlatformCapturer() Capturer {
	return &dx11Capturer{
		baseCapturer: baseCapturer{osType: "windows"},
	}
}

func (d *dx11Capturer) Start() error {
	// Stub: Initialize DX11 capture
	d.running = true
	return nil
}

func (d *dx11Capturer) Stop() {
	d.running = false
}

func (d *dx11Capturer) IsRunning() bool {
	return d.running
}

func (d *dx11Capturer) GetFrame() ([]byte, error) {
	return nil, fmt.Errorf("platform capturer not yet implemented: DX11 capture requires Windows Desktop Duplication API bindings")
}
