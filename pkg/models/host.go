package models

import "time"

// HostStatus represents the possible states of a host
type HostStatus string

const (
	HostStatusOffline      HostStatus = "offline"
	HostStatusOnline       HostStatus = "online"
	HostStatusStreaming    HostStatus = "streaming"
	HostStatusMaintenance  HostStatus = "maintenance"
)

// Host represents a gaming PC running the HelixPlay Host Agent
type Host struct {
	ID              string     `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	HardwareID      string     `json:"hardware_id" db:"hardware_id"`
	OS              string     `json:"os" db:"os"`
	OSVersion       string     `json:"os_version" db:"os_version"`
	LastSeenAt      *time.Time `json:"last_seen_at" db:"last_seen_at"`
	Status          HostStatus `json:"status" db:"status"`
	LanIP           *string    `json:"lan_ip" db:"lan_ip"`
	WanEndpoint     *string    `json:"wan_endpoint" db:"wan_endpoint"`
	DiscoveryPort   int        `json:"discovery_port" db:"discovery_port"`
	ThermalThrottle bool       `json:"thermal_throttle" db:"thermal_throttle"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}
