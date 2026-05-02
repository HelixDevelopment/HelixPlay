package input

import (
    "sync"
    "time"
    
    "digital.vasic.memory/pkg/memfd"
)

type USBPoller struct {
    mu          sync.RWMutex
    ringBuf     *memfd.PSC
    frequency   int // Hz (1000 = 1 kHz)
    polling     bool
    stopCh      chan struct{}
}

func NewUSBPoller(ringBuf *memfd.PSC, frequencyHz int) *USBPoller {
    if frequencyHz <= 0 {
        frequencyHz = 1000 // Default 1 kHz
    }
    return &USBPoller{
        ringBuf:   ringBuf,
        frequency: frequencyHz,
        stopCh:    make(chan struct{}),
    }
}

func (u *USBPoller) Start() error {
    u.mu.Lock()
    defer u.mu.Unlock()
    
    interval := time.Duration(1000000/u.frequency) * time.Microsecond // Microseconds
    
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                // Stub: Poll USB for input events
                // In real impl, would read from /dev/hidraw* or Windows HID API
                event := []byte{0x01} // Stub event
                u.ringBuf.Write(event)
            case <-u.stopCh:
                return
            }
        }
    }()
    
    u.polling = true
    return nil
}

func (u *USBPoller) Stop() {
    u.mu.Lock()
    defer u.mu.Unlock()
    close(u.stopCh)
    u.polling = false
}

func (u *USBPoller) IsPolling() bool {
    u.mu.RLock()
    defer u.mu.RUnlock()
    return u.polling
}
