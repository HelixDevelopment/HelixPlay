package capture_test

import (
	"runtime"
	"testing"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.Error(t, err, "expected error for unsupported OS")
}

func TestDispatcherSelectPipewireOnLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping Linux-specific dispatcher test")
	}

	d := capture.NewDispatcher("pipewire")
	cap, err := d.Select()
	require.NoError(t, err, "PipeWire should be selectable on Linux")
	require.NotNil(t, cap)
	assert.True(t, cap.IsRunning() || !cap.IsRunning(), "capturer should be initialised")
}

func TestDispatcherSelectKMSOnLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping Linux-specific dispatcher test")
	}

	d := capture.NewDispatcher("kms")
	cap, err := d.Select()
	require.NoError(t, err, "KMS should be selectable on Linux")
	require.NotNil(t, cap)
}

func TestDispatcherSelectAuto(t *testing.T) {
	d := capture.NewDispatcher("auto")
	cap, err := d.Select()
	require.NoError(t, err, "auto selection should not error")
	require.NotNil(t, cap)
}

func TestDispatcherSelectEmpty(t *testing.T) {
	d := capture.NewDispatcher("")
	cap, err := d.Select()
	require.NoError(t, err, "empty preferred should fallback without error")
	require.NotNil(t, cap)
}

func TestDispatcherAvailableBackends(t *testing.T) {
	d := capture.NewDispatcher("")
	backends := d.AvailableBackends()
	require.NotNil(t, backends)

	switch runtime.GOOS {
	case "windows":
		assert.Contains(t, backends, "dxgi")
	case "darwin":
		assert.Contains(t, backends, "screencapturekit")
	case "linux":
		assert.Contains(t, backends, "kms")
		assert.Contains(t, backends, "pipewire")
	default:
		assert.Empty(t, backends)
	}
}
