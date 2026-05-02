package capability_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/pkg/core/capability"
)

func TestGRPCDiscovery(t *testing.T) {
    d := capability.NewDiscovery("host-agent", "localhost:50051")
    endpoints, err := d.Discover()
    if err != nil {
        t.Fatalf("Discover failed: %v", err)
    }
    if len(endpoints) == 0 {
        t.Error("Expected at least one endpoint")
    }
}

func TestDiscoveryByType(t *testing.T) {
    d := capability.NewDiscovery("host-agent", "localhost:50051")
    endpoints := d.DiscoverByType("host-agent")
    if len(endpoints) == 0 {
        t.Error("Expected at least one host-agent endpoint")
    }
}

func TestProtocolNegotiation(t *testing.T) {
    n := capability.NewNegotiator()
    clientCaps := capability.Capabilities{
        Codecs:     []string{"H.264", "HEVC"},
        Resolutions: []string{"1080p", "4K"},
        MaxFPS:     120,
    }
    serverCaps := capability.Capabilities{
        Codecs:     []string{"H.264", "HEVC", "AV1"},
        Resolutions: []string{"720p", "1080p", "1440p", "4K"},
        MaxFPS:     240,
    }
    result := n.Negotiate(clientCaps, serverCaps)
    if !result.Success {
        t.Error("Expected negotiation to succeed")
    }
    if len(result.AgreedCodecs) == 0 {
        t.Error("Expected at least one agreed codec")
    }
}

func TestProtocolFallback(t *testing.T) {
    n := capability.NewNegotiator()
    clientCaps := capability.Capabilities{
        Codecs: []string{"VP8"}, // Not supported by server
    }
    serverCaps := capability.Capabilities{
        Codecs: []string{"H.264", "HEVC"},
    }
    result := n.Negotiate(clientCaps, serverCaps)
    if result.Success {
        t.Error("Expected negotiation to fail")
    }
}
