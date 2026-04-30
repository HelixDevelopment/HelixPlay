package discovery

import (
    "fmt"
    "net"
    "time"
)

type Beacon struct {
    serviceName string
    serviceType string
    port        int
    running     bool
    stopCh       chan struct{}
}

func NewBeacon(serviceName, serviceType string, port int) (*Beacon, error) {
    if port <= 0 || port > 65535 {
        return nil, fmt.Errorf("invalid port: %d", port)
    }
    return &Beacon{
        serviceName: serviceName,
        serviceType: serviceType,
        port:        port,
        stopCh:       make(chan struct{}),
    }, nil
}

func (b *Beacon) Start() error {
    go b.announceLoop()
    b.running = true
    return nil
}

func (b *Beacon) announceLoop() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            conn, _ := net.DialUDP("udp", nil, &net.UDPAddr{
                IP:   net.ParseIP("224.0.0.251"),
                Port: 5353,
            })
            if conn != nil {
                defer conn.Close()
                payload := []byte(fmt.Sprintf("%s._%s._tcp.local", b.serviceName, b.serviceType))
                conn.Write(payload)
            }
        case <-b.stopCh:
            return
        }
    }
}

func (b *Beacon) Stop() {
    close(b.stopCh)
    b.running = false
}

func (b *Beacon) IsRunning() bool {
    return b.running
}
