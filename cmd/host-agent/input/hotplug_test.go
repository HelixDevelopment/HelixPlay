package input_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/input"
)

func TestHotplugDetect(t *testing.T) {
    hp := input.NewHotPlug()
    // Simulate controller connected
    hp.ControllerConnected("DualSense-001")
    if !hp.IsConnected("DualSense-001") {
        t.Error("Expected DualSense-001 to be connected")
    }
}

func TestHotplugDisconnect(t *testing.T) {
    hp := input.NewHotPlug()
    hp.ControllerConnected("DualSense-001")
    hp.ControllerDisconnected("DualSense-001")
    if hp.IsConnected("DualSense-001") {
        t.Error("Expected DualSense-001 to be disconnected")
    }
}

func TestRenegotiateCapability(t *testing.T) {
    hp := input.NewHotPlug()
    hp.ControllerConnected("DualSense-001")
    
    // Initial capabilities
    caps := hp.GetCapabilites("DualSense-001")
    if caps == nil {
        t.Fatal("Expected non-nil capabilities")
    }
    
    // Simulate mid-session capability change (e.g., firmware update, new features)
    hp.UpdateCapability("DualSense-001", "haptics", "enhanced")
    caps = hp.GetCapabilites("DualSense-001")
    if caps["haptics"] != "enhanced" {
        t.Errorf("Expected enhanced haptics, got %s", caps["haptics"])
    }
}

func TestListConnected(t *testing.T) {
    hp := input.NewHotPlug()
    hp.ControllerConnected("DualSense-001")
    hp.ControllerConnected("Xbox-002")
    
    connected := hp.ListConnected()
    if len(connected) != 2 {
        t.Errorf("Expected 2 connected controllers, got %d", len(connected))
    }
}
