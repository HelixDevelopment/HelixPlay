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
	// Stub: Initialize ScreenCaptureKit
	s.running = true
	return nil
}

func (s *screenCaptureKitCapturer) Stop() {
	s.running = false
}

func (s *screenCaptureKitCapturer) IsRunning() bool {
	return s.running
}

func (s *screenCaptureKitCapturer) GetFrame() ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}
