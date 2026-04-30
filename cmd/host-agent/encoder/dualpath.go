package encoder

import (
    "sync"
)

type DualPath struct {
    mu           sync.RWMutex
    encoder      HardwareEncoder
    streaming    bool
    recording    bool
    streamOutput []byte
    recordOutput []byte
}

func NewDualPath(encoderType string) *DualPath {
    enc := NewHardwareEncoder(encoderType)
    if enc == nil {
        return nil
    }
    return &DualPath{
        encoder: enc,
    }
}

func (d *DualPath) Start() error {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.streaming = true
    d.recording = true
    // Stub: Start stream and record encoders
    d.streamOutput = []byte{0x01, 0x02, 0x03}
    d.recordOutput = []byte{0x04, 0x05, 0x06}
    return nil
}

func (d *DualPath) Stop() {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.streaming = false
    d.recording = false
    d.encoder.Close()
}

func (d *DualPath) IsStreaming() bool {
    d.mu.RLock()
    defer d.mu.RUnlock()
    return d.streaming
}

func (d *DualPath) IsRecording() bool {
    d.mu.RLock()
    defer d.mu.RUnlock()
    return d.recording
}

func (d *DualPath) GetStreamOutput() []byte {
    d.mu.RLock()
    defer d.mu.RUnlock()
    return d.streamOutput
}

func (d *DualPath) GetRecordOutput() []byte {
    d.mu.RLock()
    defer d.mu.RUnlock()
    return d.recordOutput
}
