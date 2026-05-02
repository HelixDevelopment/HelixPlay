package session

import (
	"testing"
)

func TestLoadBalancerSelect(t *testing.T) {
	lb := NewLoadBalancer()
	candidates := []*HostCandidate{
		{ID: "h1", Address: "10.0.0.1", Codecs: []string{"h264"}, ActiveSessions: 0, MaxSessions: 4},
		{ID: "h2", Address: "10.0.0.2", Codecs: []string{"h264", "hevc"}, ActiveSessions: 2, MaxSessions: 4},
		{ID: "h3", Address: "10.0.0.3", Codecs: []string{"av1"}, ActiveSessions: 0, MaxSessions: 4},
	}

	lb.RecordLatency("h1", 20)
	lb.RecordLatency("h2", 30)
	lb.RecordLatency("h3", 15)

	selected, err := lb.Select(candidates, "h264")
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if selected.ID != "h1" {
		t.Errorf("expected h1 (lowest latency h264), got %s", selected.ID)
	}
}

func TestLoadBalancerSelectWithCodecFilter(t *testing.T) {
	lb := NewLoadBalancer()
	candidates := []*HostCandidate{
		{ID: "h1", Codecs: []string{"h264"}, ActiveSessions: 0},
		{ID: "h2", Codecs: []string{"hevc"}, ActiveSessions: 0},
	}

	selected, err := lb.Select(candidates, "hevc")
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if selected.ID != "h2" {
		t.Errorf("expected h2, got %s", selected.ID)
	}

	_, err = lb.Select(candidates, "av1")
	if err == nil {
		t.Fatal("expected error for unsupported codec")
	}
}

func TestLoadBalancerSelectEmpty(t *testing.T) {
	lb := NewLoadBalancer()
	_, err := lb.Select([]*HostCandidate{}, "h264")
	if err == nil {
		t.Fatal("expected error for empty candidates")
	}
}

func TestLoadBalancerSessionPenalty(t *testing.T) {
	lb := NewLoadBalancer()
	candidates := []*HostCandidate{
		{ID: "h1", Codecs: []string{"h264"}, ActiveSessions: 3, MaxSessions: 4},
		{ID: "h2", Codecs: []string{"h264"}, ActiveSessions: 0, MaxSessions: 4},
	}

	lb.RecordLatency("h1", 10)
	lb.RecordLatency("h2", 35)

	// h1 has lower latency but more sessions; h2 should be preferred
	selected, err := lb.Select(candidates, "h264")
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if selected.ID != "h2" {
		t.Errorf("expected h2 (better session availability), got %s", selected.ID)
	}
}
