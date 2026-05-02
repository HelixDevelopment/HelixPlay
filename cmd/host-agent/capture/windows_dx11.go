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
	return fmt.Errorf("dx11 capture unavailable: requires Windows Desktop Duplication API bindings (not available in this build)")
}

func (d *dx11Capturer) Stop() {
	d.running = false
}

func (d *dx11Capturer) IsRunning() bool {
	return d.running
}

func (d *dx11Capturer) GetFrame() ([]byte, error) {
	if !d.running {
		return nil, fmt.Errorf("capturer not running")
	}
	return nil, fmt.Errorf("dx11 capture requires Windows Desktop Duplication API bindings")
}
