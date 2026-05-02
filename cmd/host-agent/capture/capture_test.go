package capture_test

import (
	"runtime"
	"testing"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capture"
)

func TestCaptureStartStopOnCurrentPlatform(t *testing.T) {
	cap := capture.NewCapturer(runtime.GOOS)
	err := cap.Start()
	if runtime.GOOS == "linux" {
		// On Linux without PipeWire, Start should return an error
		if err == nil {
			// If pipewire is available, verify GetFrame returns the expected error
			if cap.IsRunning() {
				_, frameErr := cap.GetFrame()
				if frameErr == nil {
					t.Fatal("expected GetFrame to return error for stub capturer")
				}
			}
		} else {
			// Expected error when pipewire is not available
			t.Logf("Capture start returned expected error: %v", err)
		}
	} else {
		// Darwin/Windows should return errors since native bindings aren't available
		if err == nil {
			t.Fatal("expected Start to return error on non-Linux or when native bindings unavailable")
		}
	}
}

func TestCaptureUnsupportedOS(t *testing.T) {
	cap := capture.NewCapturer("unsupported")
	err := cap.Start()
	if err == nil {
		t.Fatal("Expected error for unsupported OS")
	}
}
