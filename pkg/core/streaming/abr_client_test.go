package streaming

import (
	"testing"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/network"
)

func TestABRClientDefaultProfiles(t *testing.T) {
	profiles := DefaultProfiles()
	if len(profiles) == 0 {
		t.Fatal("expected default profiles")
	}
	if profiles[0].BitrateKbps <= 0 {
		t.Fatal("expected positive bitrate")
	}
}

func TestABRClientRecommendStepDown(t *testing.T) {
	est := network.NewEstimator(10)
	// Simulate poor network
	for i := 0; i < 5; i++ {
		est.Record(200, 5000, 5.0)
	}

	client := NewABRClient(est, DefaultProfiles())
	initial := client.CurrentProfile()

	profile, changed, err := client.Recommend()
	if err != nil {
		t.Fatalf("Recommend failed: %v", err)
	}
	if !changed {
		t.Fatal("expected recommendation to change on poor network")
	}
	if profile.BitrateKbps >= initial.BitrateKbps {
		t.Fatal("expected lower bitrate profile")
	}
}

func TestABRClientRecommendStepUp(t *testing.T) {
	est := network.NewEstimator(10)
	// Simulate excellent network
	for i := 0; i < 10; i++ {
		est.Record(15, 60000, 0.0)
	}

	client := NewABRClient(est, DefaultProfiles())
	// Force to a lower profile first
	client.ForceProfile("480p30")
	initial := client.CurrentProfile()

	profile, changed, err := client.Recommend()
	if err != nil {
		t.Fatalf("Recommend failed: %v", err)
	}
	if !changed {
		t.Fatal("expected recommendation to change on good network")
	}
	if profile.BitrateKbps <= initial.BitrateKbps {
		t.Fatal("expected higher bitrate profile")
	}
}

func TestABRClientNoChange(t *testing.T) {
	est := network.NewEstimator(10)
	est.Record(50, 15000, 0.5)

	client := NewABRClient(est, DefaultProfiles())
	initial := client.CurrentProfile()

	profile, changed, err := client.Recommend()
	if err != nil {
		t.Fatalf("Recommend failed: %v", err)
	}
	if changed {
		t.Fatal("expected no change with limited data")
	}
	if profile.Name != initial.Name {
		t.Fatal("expected same profile")
	}
}

func TestABRClientForceProfile(t *testing.T) {
	client := NewABRClient(nil, DefaultProfiles())
	if err := client.ForceProfile("1080p60"); err != nil {
		t.Fatalf("ForceProfile failed: %v", err)
	}
	if p := client.CurrentProfile(); p.Name != "1080p60" {
		t.Errorf("expected 1080p60, got %s", p.Name)
	}

	if err := client.ForceProfile("nonexistent"); err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestABRClientLadder(t *testing.T) {
	client := NewABRClient(nil, DefaultProfiles())
	ladder := client.Ladder()
	if len(ladder) != len(DefaultProfiles()) {
		t.Fatalf("expected %d profiles, got %d", len(DefaultProfiles()), len(ladder))
	}
}
