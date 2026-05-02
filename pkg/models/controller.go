package models

import "time"

// ControllerDeviceType represents supported controller types
type ControllerDeviceType string

const (
	ControllerDeviceDualSense ControllerDeviceType = "dualsense"
	ControllerDeviceXbox      ControllerDeviceType = "xbox"
	ControllerDeviceGeneric   ControllerDeviceType = "generic"
)

// ControllerConnectionType represents physical connection methods
type ControllerConnectionType string

const (
	ControllerConnUSB        ControllerConnectionType = "usb"
	ControllerConnDongle24   ControllerConnectionType = "dongle_2_4ghz"
	ControllerConnBluetooth  ControllerConnectionType = "bluetooth"
)

// Controller represents input device state forwarded from client to host
type Controller struct {
	ID                   string                 `json:"id" db:"id"`
	SessionID            string                 `json:"session_id" db:"session_id"`
	DeviceType           ControllerDeviceType   `json:"device_type" db:"device_type"`
	ConnectionType       ControllerConnectionType `json:"connection_type" db:"connection_type"`
	HapticsEnabled       bool                   `json:"haptics_enabled" db:"haptics_enabled"`
	AdaptiveTriggerEnabled bool                 `json:"adaptive_trigger_enabled" db:"adaptive_trigger_enabled"`
	GyroEnabled          bool                   `json:"gyro_enabled" db:"gyro_enabled"`
	AudioJackEnabled     bool                   `json:"audio_jack_enabled" db:"audio_jack_enabled"`
	PollRateHz           int                    `json:"poll_rate_hz" db:"poll_rate_hz"`
	LatencyUs            *int                   `json:"latency_us" db:"latency_us"`
	ConnectedAt          time.Time              `json:"connected_at" db:"connected_at"`
	DisconnectedAt       *time.Time             `json:"disconnected_at" db:"disconnected_at"`
}
