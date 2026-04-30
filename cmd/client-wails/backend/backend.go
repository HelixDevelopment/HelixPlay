package backend

import (
    "sync"
)

type Backend struct {
    mu        sync.RWMutex
    running   bool
    connected bool
    address   string
}

func NewBackend() *Backend {
    return &Backend{}
}

func (b *Backend) Start() error {
    b.mu.Lock()
    defer b.mu.Unlock()
    // Stub: Initialize Wails backend
    b.running = true
    return nil
}

func (b *Backend) Stop() {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.running = false
    b.connected = false
}

func (b *Backend) IsRunning() bool {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.running
}

func (b *Backend) Connect(address string) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    // Stub: Connect to HelixPlay host
    b.address = address
    b.connected = true
    return nil
}

func (b *Backend) IsConnected() bool {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.connected
}
