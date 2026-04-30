package transport

import (
    "fmt"
    "sync"
)

type QUIC struct {
    mu         sync.RWMutex
    addr       string
    connected  bool
}

func NewQUIC(addr string) (*QUIC, error) {
    if addr == "" {
        return nil, fmt.Errorf("address must be specified")
    }
    return &QUIC{
        addr:      addr,
        connected: false,
    }, nil
}

func (q *QUIC) Start() error {
    q.mu.Lock()
    defer q.mu.Unlock()
    // Stub: Initialize quic-go with RFC 9221 datagram support
    q.connected = true
    return nil
}

func (q *QUIC) Stop() {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.connected = false
}

func (q *QUIC) IsConnected() bool {
    q.mu.RLock()
    defer q.mu.RUnlock()
    return q.connected
}

func (q *QUIC) SendDatagram(data []byte) error {
    q.mu.RLock()
    defer q.mu.RUnlock()
    if !q.connected {
        return fmt.Errorf("not connected")
    }
    // Stub: Send RFC 9221 datagram
    _ = data // In real impl, would send via quic-go
    return nil
}

func (q *QUIC) ReceiveDatagram() ([]byte, error) {
    q.mu.RLock()
    defer q.mu.RUnlock()
    if !q.connected {
        return nil, fmt.Errorf("not connected")
    }
    // Stub: Receive RFC 9221 datagram
    return []byte{0x01, 0x02, 0x03}, nil
}
