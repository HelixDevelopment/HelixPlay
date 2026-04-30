package discovery_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/discovery"
)

func TestBeaconStart(t *testing.T) {
    b, err := discovery.NewBeacon("HelixPlay-Host", "tcp", 9090)
    if err != nil {
        t.Fatalf("Failed to create beacon: %v", err)
    }
    err = b.Start()
    if err != nil {
        t.Fatalf("Failed to start beacon: %v", err)
    }
    defer b.Stop()
    if !b.IsRunning() {
        t.Error("Expected beacon to be running")
    }
}
