package input_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/input"
)

func TestDualSenseConnect(t *testing.T) {
    ds := input.NewDualSense()
    if err := ds.Connect(); err != nil {
        t.Fatalf("Connect failed: %v", err)
    }
    if !ds.IsConnected() {
        t.Error("Expected DualSense to be connected")
    }
}

func TestDualSenseHaptics(t *testing.T) {
    ds := input.NewDualSense()
    ds.Connect()
    err := ds.SetHaptics(0.5, 0.8)
    if err != nil {
        t.Fatalf("SetHaptics failed: %v", err)
    }
}

func TestAdaptiveTriggers(t *testing.T) {
    ds := input.NewDualSense()
    ds.Connect()
    err := ds.SetTriggerResistance("L2", 0.7)
    if err != nil {
        t.Fatalf("SetTriggerResistance failed: %v", err)
    }
    resistance := ds.GetTriggerResistance("L2")
    if resistance != 0.7 {
        t.Errorf("Expected 0.7 resistance, got %f", resistance)
    }
}

func TestGyroAccel(t *testing.T) {
    ds := input.NewDualSense()
    ds.Connect()
    gyro, accel := ds.ReadIMU()
    if len(gyro) != 3 {
        t.Errorf("Expected 3 gyro axes, got %d", len(gyro))
    }
    if len(accel) != 3 {
        t.Errorf("Expected 3 accel axes, got %d", len(accel))
    }
}

func TestDualSenseDisconnect(t *testing.T) {
    ds := input.NewDualSense()
    ds.Connect()
    ds.Disconnect()
    if ds.IsConnected() {
        t.Error("Expected DualSense to be disconnected")
    }
}
