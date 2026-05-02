package transport

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/andybalholm/brotli"
)

type WebRTC struct {
	mu        sync.RWMutex
	encoder   string
	codec     string
	connected bool
}

func NewWebRTC(encoderType, codec string) (*WebRTC, error) {
	if encoderType == "" || codec == "" {
		return nil, fmt.Errorf("encoder and codec must be specified")
	}
	return &WebRTC{
		encoder:   encoderType,
		codec:     codec,
		connected: false,
	}, nil
}

func (w *WebRTC) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return fmt.Errorf("WebRTC transport unavailable: requires Pion v4 with DTLS 1.2 (not available in this build)")
}

func (w *WebRTC) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.connected = false
}

func (w *WebRTC) IsConnected() bool {
	w.mu.RLock()
	w.mu.RUnlock()
	return w.connected
}

func (w *WebRTC) GetCodec() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.codec
}

// Compress performs real Brotli compression on the data.
func (w *WebRTC) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	enc := brotli.NewWriter(&buf)
	if _, err := enc.Write(data); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decompress performs real Brotli decompression on the data.
func (w *WebRTC) Decompress(data []byte) ([]byte, error) {
	r := brotli.NewReader(bytes.NewReader(data))
	return io.ReadAll(r)
}
