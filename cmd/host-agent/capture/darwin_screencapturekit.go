//go:build darwin

package capture

import "fmt"

type screenCaptureKitCapturer struct {
	baseCapturer
}

func newPlatformCapturer() Capturer {
	return &screenCaptureKitCapturer{
		baseCapturer: baseCapturer{osType: "darwin"},
	}
}

func (s *screenCaptureKitCapturer) Start() error {
	return fmt.Errorf("screencapturekit capture unavailable: requires macOS 12.3+ with CGo/objc bridge bindings (not available in this build)")
}

func (s *screenCaptureKitCapturer) Stop() {
	s.running = false
}

func (s *screenCaptureKitCapturer) IsRunning() bool {
	return s.running
}

func (s *screenCaptureKitCapturer) GetFrame() ([]byte, error) {
	if !s.running {
		return nil, fmt.Errorf("capturer not running")
	}
	return nil, fmt.Errorf("screencapturekit capture requires macOS 12.3+ API bindings (CGo/objc bridge)")
}
