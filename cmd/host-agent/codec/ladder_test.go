package codec_test

import (
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/codec"
	"testing"
)

func TestCodecLadder(t *testing.T) {
	ladder := codec.NewLadder()
	profiles := ladder.GetProfiles()
	if len(profiles) < 3 {
		t.Error("Expected at least 3 codec profiles")
	}
	// Check H.264 720p is present with correct bitrate
	found := false
	for _, p := range profiles {
		if p.Name == "H.264" && p.Resolution == "720p" {
			found = true
			if p.MaxBitrate != 5000 {
				t.Errorf("Expected H.264 720p max bitrate 5000, got %d", p.MaxBitrate)
			}
		}
	}
	if !found {
		t.Error("H.264 720p not found in codec ladder")
	}
}

func TestGetProfileForBandwidth(t *testing.T) {
	ladder := codec.NewLadder()

	// Low bandwidth -> H.264 720p (first profile, lowest bitrate)
	profile := ladder.GetProfileForBandwidth(2000)
	if profile.Name != "H.264" || profile.Resolution != "720p" {
		t.Errorf("Expected H.264 720p for low bandwidth, got %s %s", profile.Name, profile.Resolution)
	}

	// Medium bandwidth -> HEVC 1080p (20000 kbps)
	profile = ladder.GetProfileForBandwidth(25000)
	if profile.Name != "HEVC" || profile.Resolution != "1080p" {
		t.Errorf("Expected HEVC 1080p for medium bandwidth, got %s %s", profile.Name, profile.Resolution)
	}

	// High bandwidth -> HEVC 4K (requires 50000 kbps)
	profile = ladder.GetProfileForBandwidth(50000)
	if profile.Name != "HEVC" || profile.Resolution != "4K" {
		t.Errorf("Expected HEVC 4K for high bandwidth, got %s %s", profile.Name, profile.Resolution)
	}
}
