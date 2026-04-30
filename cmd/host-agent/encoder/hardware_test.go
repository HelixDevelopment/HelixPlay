package encoder_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
)

func TestDetectNVENC(t *testing.T) {
    enc := encoder.NewHardwareEncoder("nvenc")
    if enc == nil {
        t.Fatal("Expected NVENC encoder")
    }
    if enc.Name() != "NVENC" {
        t.Errorf("Expected name NVENC, got %s", enc.Name())
    }
}

func TestDetectVAAPI(t *testing.T) {
    enc := encoder.NewHardwareEncoder("vaapi")
    if enc == nil {
        t.Fatal("Expected VAAPI encoder")
    }
}

func TestUnsupportedEncoder(t *testing.T) {
    enc := encoder.NewHardwareEncoder("unsupported")
    if enc != nil {
        t.Error("Expected nil for unsupported encoder")
    }
}

func TestEncodeFrame(t *testing.T) {
    enc := encoder.NewHardwareEncoder("nvenc")
    frame := []byte{0x00, 0x01, 0x02}
    result, err := enc.Encode(frame)
    if err != nil {
        t.Fatalf("Encode failed: %v", err)
    }
    if len(result) == 0 {
        t.Error("Expected non-empty encoded frame")
    }
}
