// Package streaming provides adaptive bitrate (ABR) policy for streaming sessions.
package streaming

import (
	"fmt"
	"sync"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/network"
)

// ABRProfile defines a bitrate ladder step.
type ABRProfile struct {
	Name       string
	BitrateKbps int32
	Resolution string
	FPS        int32
}

// DefaultProfiles returns the standard ABR ladder.
func DefaultProfiles() []ABRProfile {
	return []ABRProfile{
		{Name: "4K60", BitrateKbps: 40000, Resolution: "4K", FPS: 60},
		{Name: "1440p60", BitrateKbps: 25000, Resolution: "1440p", FPS: 60},
		{Name: "1080p60", BitrateKbps: 15000, Resolution: "1080p", FPS: 60},
		{Name: "720p60", BitrateKbps: 8000, Resolution: "720p", FPS: 60},
		{Name: "720p30", BitrateKbps: 5000, Resolution: "720p", FPS: 30},
		{Name: "480p30", BitrateKbps: 2500, Resolution: "480p", FPS: 30},
	}
}

// ABRClient selects the best profile based on network conditions.
type ABRClient struct {
	mu       sync.RWMutex
	estimator *network.Estimator
	profiles []ABRProfile
	current  int // index into profiles
}

// NewABRClient creates an ABR client with the given estimator and profile ladder.
func NewABRClient(estimator *network.Estimator, profiles []ABRProfile) *ABRClient {
	if len(profiles) == 0 {
		profiles = DefaultProfiles()
	}
	return &ABRClient{
		estimator: estimator,
		profiles:  profiles,
		current:   len(profiles) / 2, // Start in the middle
	}
}

// CurrentProfile returns the currently selected profile.
func (a *ABRClient) CurrentProfile() ABRProfile {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.profiles[a.current]
}

// Recommend evaluates network conditions and returns a profile recommendation.
// It returns true if a change is suggested.
func (a *ABRClient) Recommend() (ABRProfile, bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	current := a.profiles[a.current]

	// Require stable network before stepping up
	stable := a.estimator.Stable()
	rec := a.estimator.Recommendation()

	// If network is poor, step down aggressively
	if rec == "poor" {
		if a.current < len(a.profiles)-1 {
			a.current++
			return a.profiles[a.current], true, nil
		}
		return current, false, nil
	}

	// If network is good and stable, step up conservatively
	if rec == "good" && stable {
		bw := a.estimator.Bandwidth()
		// Require 20% headroom above target bitrate
		headroom := float64(current.BitrateKbps) * 1.2
		if a.current > 0 && bw >= headroom {
			a.current--
			return a.profiles[a.current], true, nil
		}
	}

	return current, false, nil
}

// ForceProfile manually sets the profile by name.
func (a *ABRClient) ForceProfile(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, p := range a.profiles {
		if p.Name == name {
			a.current = i
			return nil
		}
	}
	return fmt.Errorf("profile %s not found", name)
}

// Ladder returns the full profile list.
func (a *ABRClient) Ladder() []ABRProfile {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]ABRProfile, len(a.profiles))
	copy(out, a.profiles)
	return out
}
