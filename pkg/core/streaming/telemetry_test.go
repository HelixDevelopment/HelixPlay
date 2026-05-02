package streaming

import (
	"testing"
	"time"
)

func TestTelemetryBuilder(t *testing.T) {
	b := NewTelemetryBuilder("sess-123")
	time.Sleep(10 * time.Millisecond)

	snap := b.Build(
		100, 2,
		15000,
		15.0, 45.0,
		0.1, 50.0,
		"1080p", "H.264", "dualsense",
		5,
		false,
	)

	if snap.SessionID != "sess-123" {
		t.Errorf("expected sess-123, got %s", snap.SessionID)
	}
	if snap.FramesEncoded != 100 {
		t.Errorf("expected 100 frames, got %d", snap.FramesEncoded)
	}
	if snap.BitrateKbps != 15000 {
		t.Errorf("expected 15000 kbps, got %d", snap.BitrateKbps)
	}
	if snap.DurationMs < 1 {
		t.Error("expected positive duration")
	}
}

func TestTelemetryValidate(t *testing.T) {
	valid := TelemetrySnapshot{
		SessionID: "sess-1",
		Timestamp: time.Now().UTC(),
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid: %v", err)
	}

	invalid := TelemetrySnapshot{BitrateKbps: -1}
	if err := invalid.Validate(); err == nil {
		t.Fatal("expected error for missing session_id")
	}
}

func TestTelemetryJSON(t *testing.T) {
	snap := TelemetrySnapshot{
		SessionID:    "sess-1",
		Timestamp:    time.Now().UTC(),
		BitrateKbps:  10000,
		Resolution:   "1080p",
		Codec:        "H.264",
		LatencyP50Ms: 20.0,
		LatencyP99Ms: 50.0,
	}
	data, err := snap.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}
