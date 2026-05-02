package models

import "time"

// GPUVendor represents supported GPU vendors
type GPUVendor string

const (
	GPUVendorNVIDIA GPUVendor = "nvidia"
	GPUVendorIntel  GPUVendor = "intel"
	GPUVendorAMD    GPUVendor = "amd"
	GPUVendorApple  GPUVendor = "apple"
)

// GPU represents per-GPU capabilities advertised by the Host Agent
type GPU struct {
	ID              string    `json:"id" db:"id"`
	HostID          string    `json:"host_id" db:"host_id"`
	Vendor          GPUVendor `json:"vendor" db:"vendor"`
	Model           string    `json:"model" db:"model"`
	DriverVersion   string    `json:"driver_version" db:"driver_version"`
	VRAMGB          int       `json:"vram_gb" db:"vram_gb"`
	Encoder         string    `json:"encoder" db:"encoder"`
	CodecsSupported []string  `json:"codecs_supported" db:"codecs_supported"`
	MaxResolution   string    `json:"max_resolution" db:"max_resolution"`
	MaxRefreshHz    int       `json:"max_refresh_hz" db:"max_refresh_hz"`
	NVENCSessions   *int      `json:"nvenc_sessions" db:"nvenc_sessions"`
	ThermalC        *int      `json:"thermal_c" db:"thermal_c"`
	SessionCount    int       `json:"session_count" db:"session_count"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
