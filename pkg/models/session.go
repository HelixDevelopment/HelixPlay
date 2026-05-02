package models

import "time"

// SessionStatus represents the possible states of a streaming session
type SessionStatus string

const (
	SessionStatusConnecting    SessionStatus = "connecting"
	SessionStatusNegotiating   SessionStatus = "negotiating"
	SessionStatusStreaming     SessionStatus = "streaming"
	SessionStatusPaused        SessionStatus = "paused"
	SessionStatusReconnecting  SessionStatus = "reconnecting"
	SessionStatusDisconnected  SessionStatus = "disconnected"
	SessionStatusEnded         SessionStatus = "ended"
)

// Session represents an active streaming session between a Client and a Host
type Session struct {
	ID              string        `json:"id" db:"id"`
	UserID          string        `json:"user_id" db:"user_id"`
	HostID          string        `json:"host_id" db:"host_id"`
	GameID          *string       `json:"game_id" db:"game_id"`
	GPUID           *string       `json:"gpu_id" db:"gpu_id"`
	Status          SessionStatus `json:"status" db:"status"`
	Codec           string        `json:"codec" db:"codec"`
	Transport       string        `json:"transport" db:"transport"`
	Resolution      string        `json:"resolution" db:"resolution"`
	RefreshHz       int           `json:"refresh_hz" db:"refresh_hz"`
	LatencyP50Ms    *float64      `json:"latency_p50_ms" db:"latency_p50_ms"`
	LatencyP99Ms    *float64      `json:"latency_p99_ms" db:"latency_p99_ms"`
	LatencyP999Ms   *float64      `json:"latency_p999_ms" db:"latency_p999_ms"`
	BandwidthMbps   *float64      `json:"bandwidth_mbps" db:"bandwidth_mbps"`
	PacketLossPct   *float64      `json:"packet_loss_pct" db:"packet_loss_pct"`
	ControllerType  string        `json:"controller_type" db:"controller_type"`
	RecordingEnabled bool         `json:"recording_enabled" db:"recording_enabled"`
	RecordingPath   *string       `json:"recording_path" db:"recording_path"`
	StartedAt       time.Time     `json:"started_at" db:"started_at"`
	EndedAt         *time.Time    `json:"ended_at" db:"ended_at"`
	LastActivityAt  time.Time     `json:"last_activity_at" db:"last_activity_at"`
}
