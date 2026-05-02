package encoder_test

import (
	"bytes"
	"testing"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
)

func TestSoftwareEncoderExists(t *testing.T) {
	enc := encoder.NewHardwareEncoder("software")
	if enc == nil {
		t.Fatal("Expected software encoder to always be available")
	}
	if enc.Name() != "Software" {
		t.Errorf("Expected name Software, got %s", enc.Name())
	}
}

func TestHardwareEncoderRequiresGPU(t *testing.T) {
	// Hardware encoders should return nil when no GPU is present
	// (this is the expected behavior on CI without GPUs)
	nvenc := encoder.NewHardwareEncoder("nvenc")
	if nvenc != nil {
		// If a GPU IS present, verify it returns an error for encoding
		_, err := nvenc.Encode([]byte{0x00, 0x01, 0x02})
		if err == nil {
			t.Fatal("expected NVENC encode to return error when library not loaded")
		}
	}
}

func TestUnsupportedEncoder(t *testing.T) {
	enc := encoder.NewHardwareEncoder("unsupported")
	if enc != nil {
		t.Error("Expected nil for unsupported encoder")
	}
}

func TestSoftwareEncoderRealTransform(t *testing.T) {
	enc := encoder.NewHardwareEncoder("software")
	if enc == nil {
		t.Fatal("software encoder must be available")
	}
	frame := []byte{0x00, 0x01, 0x02, 0x03, 0x03, 0x03}
	result, err := enc.Encode(frame)
	if err != nil {
		t.Fatalf("Software encode failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("Expected non-empty encoded frame")
	}
	// Verify the result is different from input (real transformation occurred)
	if bytes.Equal(result, frame) {
		t.Fatal("software encoder produced identical output - no real transformation")
	}
	// Verify software marker byte is present
	if result[0] != 0x06 {
		t.Fatalf("expected software marker 0x06, got 0x%02x", result[0])
	}
}

func TestSoftwareEncoderEmptyFrame(t *testing.T) {
	enc := encoder.NewHardwareEncoder("software")
	if enc == nil {
		t.Fatal("software encoder must be available")
	}
	result, err := enc.Encode([]byte{})
	if err != nil {
		t.Fatalf("Empty frame encode failed: %v", err)
	}
	if len(result) != 1 || result[0] != 0x00 {
		t.Fatalf("expected [0x00] for empty frame, got %v", result)
	}
}
