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
	return fmt.Errorf("QUIC transport unavailable: requires quic-go with RFC 9221 datagram support (not available in this build)")
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
	return fmt.Errorf("QUIC datagram transport not implemented")
}

func (q *QUIC) ReceiveDatagram() ([]byte, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if !q.connected {
		return nil, fmt.Errorf("not connected")
	}
	return nil, fmt.Errorf("QUIC datagram transport not implemented")
}
