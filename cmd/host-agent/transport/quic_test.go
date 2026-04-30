package transport_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
)

func TestQUICStart(t *testing.T) {
    q, err := transport.NewQUIC("localhost:4242")
    if err != nil {
        t.Fatalf("NewQUIC failed: %v", err)
    }
    if err := q.Start(); err != nil {
        t.Fatalf("Start failed: %v", err)
    }
    if !q.IsConnected() {
        t.Error("Expected QUIC to be connected")
    }
}

func TestQUICStop(t *testing.T) {
    q, _ := transport.NewQUIC("localhost:4242")
    q.Start()
    q.Stop()
    if q.IsConnected() {
        t.Error("Expected QUIC to be disconnected")
    }
}

func TestQUICSendDatagram(t *testing.T) {
    q, _ := transport.NewQUIC("localhost:4242")
    q.Start()
    data := []byte("Hello QUIC")
    err := q.SendDatagram(data)
    if err != nil {
        t.Fatalf("SendDatagram failed: %v", err)
    }
}

func TestQUICReceiveDatagram(t *testing.T) {
    q, _ := transport.NewQUIC("localhost:4242")
    q.Start()
    // Stub: receive datagram
    data, err := q.ReceiveDatagram()
    if err != nil {
        t.Fatalf("ReceiveDatagram failed: %v", err)
    }
    if len(data) == 0 {
        t.Error("Expected non-empty datagram data")
    }
}
