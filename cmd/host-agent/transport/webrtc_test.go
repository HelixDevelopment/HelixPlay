package transport_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
)

func TestWebRTCStart(t *testing.T) {
    w, err := transport.NewWebRTC("nvenc", "H.264")
    if err != nil {
        t.Fatalf("NewWebRTC failed: %v", err)
    }
    if err := w.Start(); err != nil {
        t.Fatalf("Start failed: %v", err)
    }
    if !w.IsConnected() {
        t.Error("Expected WebRTC to be connected")
    }
}

func TestWebRTCStop(t *testing.T) {
    w, _ := transport.NewWebRTC("nvenc", "H.264")
    w.Start()
    w.Stop()
    if w.IsConnected() {
        t.Error("Expected WebRTC to be disconnected")
    }
}

func TestWebRTCGetCodec(t *testing.T) {
    w, _ := transport.NewWebRTC("nvenc", "HEVC")
    if w.GetCodec() != "HEVC" {
        t.Errorf("Expected HEVC codec, got %s", w.GetCodec())
    }
}

func TestBrotliCompression(t *testing.T) {
    w, _ := transport.NewWebRTC("nvenc", "H.264")
    data := []byte("Hello HelixPlay World")
    compressed, err := w.Compress(data)
    if err != nil {
        t.Fatalf("Compress failed: %v", err)
    }
    if len(compressed) == 0 {
        t.Error("Expected non-empty compressed data")
    }
}
