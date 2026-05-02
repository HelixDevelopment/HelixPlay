package encoder

import (
	"fmt"
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
	if d.encoder == nil {
		return fmt.Errorf("no encoder available: hardware encoder '%s' not detected or not supported", d.encoder.Name())
	}
	d.streaming = true
	d.recording = true
	return nil
}

func (d *DualPath) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.streaming = false
	d.recording = false
	if d.encoder != nil {
		d.encoder.Close()
	}
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

func (d *DualPath) EncodeFrame(frame []byte) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if !d.streaming {
		return nil, fmt.Errorf("encoder not streaming")
	}
	if d.encoder == nil {
		return nil, fmt.Errorf("no encoder configured")
	}
	return d.encoder.Encode(frame)
}
