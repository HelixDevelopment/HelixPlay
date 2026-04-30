package capability

import (
    "encoding/json"
    "fmt"
)

type Metadata struct {
    GPUModel       string   `json:"gpu_model"`
    CodecsSupported []string `json:"codecs_supported"`
    MaxResolution   string   `json:"max_resolution"`
    RefreshRate     int      `json:"refresh_rate"`
    NVENCSessions   int      `json:"nvenc_sessions"`
    ThermalHeadroom float64  `json:"thermal_headroom"`
}

func Advertise() (*Metadata, error) {
    return &Metadata{
        GPUModel:       "NVIDIA RTX 4080",
        CodecsSupported: []string{"H.264", "HEVC", "AV1"},
        MaxResolution:   "4K",
        RefreshRate:     120,
        NVENCSessions:   1, // C-005: 1 session per GPU
        ThermalHeadroom: 0.3, // 30% headroom
    }, nil
}

func (c *Metadata) ToJSON() string {
    data, _ := json.Marshal(c)
    return fmt.Sprintf("HelixPlay Host: %s", string(data))
}
