package transport_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
)

func TestUDPStart(t *testing.T) {
    u, err := transport.NewUDP("localhost:0")
    if err != nil {
        t.Fatalf("NewUDP failed: %v", err)
    }
    if err := u.Start(); err != nil {
        t.Fatalf("Start failed: %v", err)
    }
    if !u.IsRunning() {
        t.Error("Expected UDP to be running")
    }
    u.Stop()
}

func TestUDPStop(t *testing.T) {
    u, _ := transport.NewUDP("localhost:0")
    u.Start()
    u.Stop()
    if u.IsRunning() {
        t.Error("Expected UDP to be stopped")
    }
}

func TestUDPSendAndReceive(t *testing.T) {
    // Create receiver - listens on ephemeral port
    receiver, err := transport.NewUDP("localhost:0")
    if err != nil {
        t.Fatal(err)
    }
    if err := receiver.Start(); err != nil {
        t.Fatalf("Receiver Start failed: %v", err)
    }
    defer receiver.Stop()

    t.Logf("Receiver addr: %v", receiver.LocalAddr())

    // Create sender - connects to receiver
    sender, err := transport.NewUDP(receiver.LocalAddr().String())
    if err != nil {
        t.Fatal(err)
    }
    if err := sender.Start(); err != nil {
        t.Fatalf("Sender Start failed: %v", err)
    }
    defer sender.Stop()

    // Send data
    data := []byte("Hello UDP")
    if err := sender.SendPacket(data); err != nil {
        t.Fatalf("SendPacket failed: %v", err)
    }

    // Receive data
    received, err := receiver.ReceivePacket()
    if err != nil {
        t.Fatalf("ReceivePacket failed: %v", err)
    }
    if len(received) == 0 {
        t.Error("Expected non-empty packet data")
    }
    t.Logf("Received: %s", string(received))
}

func TestMoonlightCompat(t *testing.T) {
    u, _ := transport.NewUDP("localhost:0")
    u.EnableMoonlightCompat()
    if !u.IsMoonlightCompat() {
        t.Error("Expected Moonlight compatibility to be enabled")
    }
}
