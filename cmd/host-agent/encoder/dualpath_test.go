package encoder_test

import (
	"testing"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
)

func TestDualPathSoftwareStart(t *testing.T) {
	dp := encoder.NewDualPath("software")
	if dp == nil {
		t.Fatal("Expected dual-path encoder with software fallback")
	}
	if err := dp.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !dp.IsStreaming() {
		t.Error("Expected streaming to be active")
	}
	if !dp.IsRecording() {
		t.Error("Expected recording to be active")
	}
}

func TestDualPathSoftwareStop(t *testing.T) {
	dp := encoder.NewDualPath("software")
	if dp == nil {
		t.Fatal("Expected dual-path encoder with software fallback")
	}
	dp.Start()
	dp.Stop()
	if dp.IsStreaming() {
		t.Error("Expected streaming to be stopped")
	}
	if dp.IsRecording() {
		t.Error("Expected recording to be stopped")
	}
}

func TestDualPathSoftwareEncodeFrame(t *testing.T) {
	dp := encoder.NewDualPath("software")
	if dp == nil {
		t.Fatal("Expected dual-path encoder with software fallback")
	}
	if err := dp.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	frame := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	encoded, err := dp.EncodeFrame(frame)
	if err != nil {
		t.Fatalf("EncodeFrame failed: %v", err)
	}
	if len(encoded) == 0 {
		t.Error("Expected non-empty encoded frame")
	}
}

func TestDualPathHardwareUnavailable(t *testing.T) {
	// On CI without GPU, nvenc should return nil
	dp := encoder.NewDualPath("nvenc")
	if dp != nil {
		// If somehow a GPU is present, Start should still fail because
		// the encoder library isn't loaded
		err := dp.Start()
		if err == nil {
			_, encErr := dp.EncodeFrame([]byte{0x01})
			if encErr == nil {
				t.Fatal("expected EncodeFrame to fail when hardware library not loaded")
			}
		}
	}
}
