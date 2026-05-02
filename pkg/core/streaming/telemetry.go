// Package streaming provides telemetry reporting for active sessions.
package streaming

import (
	"encoding/json"
	"fmt"
	"time"
)

// TelemetrySnapshot holds a single telemetry report.
type TelemetrySnapshot struct {
	SessionID       string    `json:"session_id"`
	Timestamp       time.Time `json:"timestamp"`
	DurationMs      int64     `json:"duration_ms"`
	FramesEncoded   int64     `json:"frames_encoded"`
	FramesDropped   int64     `json:"frames_dropped"`
	BitrateKbps     int32     `json:"bitrate_kbps"`
	LatencyP50Ms    float64   `json:"latency_p50_ms"`
	LatencyP99Ms    float64   `json:"latency_p99_ms"`
	PacketLossPct   float64   `json:"packet_loss_pct"`
	BandwidthMbps   float64   `json:"bandwidth_mbps"`
	Resolution      string    `json:"resolution"`
	Codec           string    `json:"codec"`
	ControllerType  string    `json:"controller_type"`
	HapticEvents    int64     `json:"haptic_events"`
	ThermalThrottle bool      `json:"thermal_throttle"`
}

// Builder accumulates telemetry data and produces snapshots.
type Builder struct {
	sessionID string
	startTime time.Time
}

// NewTelemetryBuilder creates a builder for the given session.
func NewTelemetryBuilder(sessionID string) *Builder {
	return &Builder{
		sessionID: sessionID,
		startTime: time.Now().UTC(),
	}
}

// Build creates a TelemetrySnapshot from current state.
func (b *Builder) Build(
	framesEncoded, framesDropped int64,
	bitrateKbps int32,
	latencyP50, latencyP99 float64,
	packetLossPct, bandwidthMbps float64,
	resolution, codec, controllerType string,
	hapticEvents int64,
	thermalThrottle bool,
) TelemetrySnapshot {
	return TelemetrySnapshot{
		SessionID:       b.sessionID,
		Timestamp:       time.Now().UTC(),
		DurationMs:      time.Since(b.startTime).Milliseconds(),
		FramesEncoded:   framesEncoded,
		FramesDropped:   framesDropped,
		BitrateKbps:     bitrateKbps,
		LatencyP50Ms:    latencyP50,
		LatencyP99Ms:    latencyP99,
		PacketLossPct:   packetLossPct,
		BandwidthMbps:   bandwidthMbps,
		Resolution:      resolution,
		Codec:           codec,
		ControllerType:  controllerType,
		HapticEvents:    hapticEvents,
		ThermalThrottle: thermalThrottle,
	}
}

// ToJSON serializes the snapshot to JSON.
func (ts *TelemetrySnapshot) ToJSON() ([]byte, error) {
	return json.Marshal(ts)
}

// Validate checks that mandatory fields are present.
func (ts *TelemetrySnapshot) Validate() error {
	if ts.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if ts.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	if ts.BitrateKbps < 0 {
		return fmt.Errorf("bitrate cannot be negative")
	}
	return nil
}
