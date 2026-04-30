package codec_test

import (
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/codec"
	"testing"
)

func TestNegotiateCodecs(t *testing.T) {
	clientCodecs := []string{"H.264", "HEVC"}
	serverCodecs := []string{"H.264", "HEVC", "AV1"}

	neg := codec.NewNegotiator()
	result := neg.Negotiate(clientCodecs, serverCodecs)

	if len(result.Agreed) != 2 {
		t.Errorf("Expected 2 agreed codecs, got %d", len(result.Agreed))
	}

	// Check that AV1 is not in agreed (client doesn't support it)
	for _, c := range result.Agreed {
		if c == "AV1" {
			t.Error("AV1 should not be agreed (client doesn't support it)")
		}
	}
}

func TestNegotiateNoMatch(t *testing.T) {
	clientCodecs := []string{"VP8"}
	serverCodecs := []string{"H.264", "HEVC"}

	neg := codec.NewNegotiator()
	result := neg.Negotiate(clientCodecs, serverCodecs)

	if len(result.Agreed) != 0 {
		t.Errorf("Expected 0 agreed codecs, got %d", len(result.Agreed))
	}
	if !result.Failed {
		t.Error("Expected negotiation to fail")
	}
}
