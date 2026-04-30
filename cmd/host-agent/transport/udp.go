package transport

import (
    "fmt"
    "net"
    "sync"
)

type UDP struct {
    mu               sync.RWMutex
    addr             string
    conn             *net.UDPConn
    running          bool
    moonlightCompat  bool
}

func NewUDP(addr string) (*UDP, error) {
    if addr == "" {
        return nil, fmt.Errorf("address must be specified")
    }
    return &UDP{
        addr:    addr,
        running: false,
    }, nil
}

func (u *UDP) Start() error {
    u.mu.Lock()
    defer u.mu.Unlock()
    addr, err := net.ResolveUDPAddr("udp", u.addr)
    if err != nil {
        return err
    }
    // Parsec BUD style: use DialUDP for connected socket (low latency)
    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        return err
    }
    u.conn = conn
    u.running = true
    return nil
}

func (u *UDP) Stop() {
    u.mu.Lock()
    defer u.mu.Unlock()
    if u.conn != nil {
        u.conn.Close()
    }
    u.running = false
}

func (u *UDP) IsRunning() bool {
    u.mu.RLock()
    defer u.mu.RUnlock()
    return u.running
}

func (u *UDP) SendPacket(data []byte) error {
    u.mu.RLock()
    defer u.mu.RUnlock()
    if !u.running || u.conn == nil {
        return fmt.Errorf("not running")
    }
    _, err := u.conn.Write(data)
    return err
}

func (u *UDP) ReceivePacket() ([]byte, error) {
    u.mu.RLock()
    defer u.mu.RUnlock()
    if !u.running || u.conn == nil {
        return nil, fmt.Errorf("not running")
    }
    buf := make([]byte, 1500)
    n, _, err := u.conn.ReadFromUDP(buf)
    if err != nil {
        return nil, err
    }
    return buf[:n], nil
}

func (u *UDP) EnableMoonlightCompat() {
    u.mu.Lock()
    defer u.mu.Unlock()
    u.moonlightCompat = true
}

func (u *UDP) IsMoonlightCompat() bool {
    u.mu.RLock()
    defer u.mu.RUnlock()
    return u.moonlightCompat
}

func (u *UDP) LocalAddr() net.Addr {
    u.mu.RLock()
    defer u.mu.RUnlock()
    if u.conn == nil {
        return nil
    }
    return u.conn.LocalAddr()
}

func (u *UDP) SendTo(data []byte, addr *net.UDPAddr) error {
    u.mu.RLock()
    defer u.mu.RUnlock()
    if !u.running || u.conn == nil {
        return fmt.Errorf("not running")
    }
    _, err := u.conn.WriteToUDP(data, addr)
    return err
}
