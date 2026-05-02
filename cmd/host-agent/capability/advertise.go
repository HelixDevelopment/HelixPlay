package capability

import (
	"encoding/json"
	"fmt"
)

// Metadata holds the complete capability advertisement for a host.
type Metadata struct {
	GPUModel        string   `json:"gpu_model"`
	CodecsSupported []string `json:"codecs_supported"`
	MaxResolution   string   `json:"max_resolution"`
	RefreshRate     int      `json:"refresh_rate"`
	NVENCSessions   int      `json:"nvenc_sessions"`
	ThermalHeadroom float64  `json:"thermal_headroom"`
	GPUCount        int      `json:"gpu_count"`
	HardwareID      string   `json:"hardware_id,omitempty"`
}

// Advertise detects hardware and returns the host's capability metadata.
func Advertise() (*Metadata, error) {
	gpus, err := EnumerateGPUs()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate GPUs: %w", err)
	}

	if len(gpus) == 0 {
		return nil, fmt.Errorf("no GPUs detected")
	}

	primary := gpus[0]

	// Merge codecs from all GPUs
	codecSet := make(map[string]struct{})
	for _, gpu := range gpus {
		for _, c := range gpu.CodecsSupported {
			codecSet[c] = struct{}{}
		}
	}
	codecs := make([]string, 0, len(codecSet))
	for c := range codecSet {
		codecs = append(codecs, c)
	}

	hwID, _ := Fingerprint()

	return &Metadata{
		GPUModel:        primary.Model,
		CodecsSupported: codecs,
		MaxResolution:   primary.MaxResolution,
		RefreshRate:     primary.MaxRefreshHz,
		NVENCSessions:   primary.NVENCSessions * len(gpus),
		ThermalHeadroom: primary.ThermalHeadroom,
		GPUCount:        len(gpus),
		HardwareID:      hwID,
	}, nil
}

// ToJSON serializes metadata to a compact JSON string.
func (c *Metadata) ToJSON() string {
	data, _ := json.Marshal(c)
	return string(data)
}
