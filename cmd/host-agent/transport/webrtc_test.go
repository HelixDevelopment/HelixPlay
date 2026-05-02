package transport_test

import (
	"bytes"
	"testing"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
)

func TestWebRTCStartReturnsError(t *testing.T) {
	w, err := transport.NewWebRTC("software", "H.264")
	if err != nil {
		t.Fatalf("NewWebRTC failed: %v", err)
	}
	// WebRTC transport is not implemented; Start must return an error
	err = w.Start()
	if err == nil {
		t.Fatal("Expected Start to return error for unimplemented WebRTC transport")
	}
	if w.IsConnected() {
		t.Error("Expected WebRTC to not be connected after failed Start")
	}
}

func TestWebRTCGetCodec(t *testing.T) {
	w, _ := transport.NewWebRTC("software", "HEVC")
	if w.GetCodec() != "HEVC" {
		t.Errorf("Expected HEVC codec, got %s", w.GetCodec())
	}
}

func TestBrotliCompressionReal(t *testing.T) {
	w, _ := transport.NewWebRTC("software", "H.264")
	data := bytes.Repeat([]byte("Hello HelixPlay World"), 100)
	compressed, err := w.Compress(data)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	if len(compressed) == 0 {
		t.Fatal("Expected non-empty compressed data")
	}
	// Brotli should compress repetitive data to be smaller
	if len(compressed) >= len(data) {
		t.Fatalf("compression did not reduce size: %d -> %d", len(data), len(compressed))
	}
	// Verify round-trip
	decompressed, err := w.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}
	if !bytes.Equal(decompressed, data) {
		t.Fatal("decompressed data does not match original")
	}
}
