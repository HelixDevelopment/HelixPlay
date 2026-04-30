package transport_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
)

func TestABRPolicy(t *testing.T) {
    p := transport.NewPolicies()
    bitrate := p.CalculateBitrate(5000)
    if bitrate == 0 {
        t.Error("Expected non-zero bitrate")
    }
    lowBitrate := p.CalculateBitrate(500)
    if lowBitrate >= bitrate {
        t.Errorf("Expected lower bitrate for low bandwidth, got %d vs %d", lowBitrate, bitrate)
    }
}

func TestFECPolicy(t *testing.T) {
    p := transport.NewPolicies()
    p.SetNetworkCondition("unstable")
    if !p.IsFECEnabled() {
        t.Error("Expected FEC to be enabled for unstable network")
    }
    p.SetNetworkCondition("stable")
    if p.IsFECEnabled() {
        t.Error("Expected FEC to be disabled for stable network")
    }
}

func TestSQPolicies(t *testing.T) {
    p := transport.NewPolicies()
    p.SetQuality("high")
    if p.GetSQPreset() != "high" {
        t.Errorf("Expected SQP reset 'high', got '%s'", p.GetSQPreset())
    }
    p.SetQuality("low")
    if p.GetSQPreset() != "low" {
        t.Errorf("Expected SQP reset 'low', got '%s'", p.GetSQPreset())
    }
}

func TestFramePacingVRR(t *testing.T) {
    p := transport.NewPolicies()
    p.EnableVRR(120)
    if !p.IsVRREnabled() {
        t.Error("Expected VRR to be enabled")
    }
    paced := p.PaceFrame(16.67)
    if !paced {
        t.Error("Expected frame to be paced")
    }
}
