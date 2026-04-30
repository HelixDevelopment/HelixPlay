package capture_test

import (
	"testing"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capture"
)

func TestCaptureStartStop(t *testing.T) {
	cap := capture.NewCapturer("linux")
	if err := cap.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !cap.IsRunning() {
		t.Error("Expected capture to be running")
	}
	cap.Stop()
	if cap.IsRunning() {
		t.Error("Expected capture to be stopped")
	}
}

func TestCaptureUnsupportedOS(t *testing.T) {
	cap := capture.NewCapturer("unsupported")
	err := cap.Start()
	if err == nil {
		t.Error("Expected error for unsupported OS")
	}
}
