package capability

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GPUInfo holds detected GPU metadata.
type GPUInfo struct {
	Vendor           string  `json:"vendor"`
	Model            string  `json:"model"`
	DriverVersion    string  `json:"driver_version"`
	VRAMMB           int     `json:"vram_mb"`
	CodecsSupported  []string `json:"codecs_supported"`
	MaxResolution    string  `json:"max_resolution"`
	MaxRefreshHz     int     `json:"max_refresh_hz"`
	NVENCSessions    int     `json:"nvenc_sessions"`
	ThermalHeadroom  float64 `json:"thermal_headroom"`
}

// EnumerateGPUs detects all available GPUs on the host.
func EnumerateGPUs() ([]GPUInfo, error) {
	var gpus []GPUInfo

	// Try NVIDIA first
	if nvidia := detectNVIDIA(); nvidia != nil {
		gpus = append(gpus, *nvidia)
	}

	// Try AMD
	if amd := detectAMD(); amd != nil {
		gpus = append(gpus, *amd)
	}

	// Try Intel
	if intel := detectIntel(); intel != nil {
		gpus = append(gpus, *intel)
	}

	// Fallback for Apple Silicon or unknown
	if len(gpus) == 0 {
		gpus = append(gpus, fallbackGPU())
	}

	return gpus, nil
}

func detectNVIDIA() *GPUInfo {
	// Check for NVIDIA driver presence
	if _, err := os.Stat("/proc/driver/nvidia/gpus"); os.IsNotExist(err) {
		return nil
	}

	info := &GPUInfo{
		Vendor:          "NVIDIA",
		CodecsSupported: []string{"H.264", "HEVC", "AV1"},
		MaxResolution:   "4K",
		MaxRefreshHz:    240,
		NVENCSessions:   1,
		ThermalHeadroom: 0.3,
	}

	// Parse GPU information from procfs
	if data, err := os.ReadFile("/proc/driver/nvidia/gpus/0/information"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "Model:") {
				info.Model = strings.TrimSpace(line[len("Model:"):])
			}
			if strings.HasPrefix(line, "Video BIOS:") {
				info.DriverVersion = strings.TrimSpace(line[len("Video BIOS:"):])
			}
		}
	}

	// Try nvidia-smi for VRAM and temperature
	if data, err := os.ReadFile("/proc/driver/nvidia/gpus/0/registry"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "Framebuffer") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					if mb, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
						info.VRAMMB = mb / 1024 / 1024 // bytes to MB
					}
				}
			}
		}
	}

	if info.Model == "" {
		info.Model = "NVIDIA GPU"
	}
	if info.VRAMMB == 0 {
		info.VRAMMB = 8192 // conservative default
	}

	return info
}

func detectAMD() *GPUInfo {
	// Check for AMD GPU via DRM sysfs
	vendorPath := "/sys/class/drm/card0/device/vendor"
	if data, err := os.ReadFile(vendorPath); err == nil {
		vendor := strings.TrimSpace(string(data))
		if vendor == "0x1002" { // AMD PCI vendor ID
			info := &GPUInfo{
				Vendor:          "AMD",
				CodecsSupported: []string{"H.264", "HEVC"},
				MaxResolution:   "4K",
				MaxRefreshHz:    144,
				NVENCSessions:   0, // AMD uses VCE/VCN, not NVENC
				ThermalHeadroom: 0.25,
			}

			// Try to read model name
			if data, err := os.ReadFile("/sys/class/drm/card0/device/product_name"); err == nil {
				info.Model = strings.TrimSpace(string(data))
			} else if data, err := os.ReadFile("/sys/class/drm/card0/device/product_number"); err == nil {
				info.Model = "AMD " + strings.TrimSpace(string(data))
			}

			if info.Model == "" {
				info.Model = "AMD GPU"
			}
			info.VRAMMB = 8192

			return info
		}
	}
	return nil
}

func detectIntel() *GPUInfo {
	vendorPath := "/sys/class/drm/card0/device/vendor"
	if data, err := os.ReadFile(vendorPath); err == nil {
		vendor := strings.TrimSpace(string(data))
		if vendor == "0x8086" { // Intel PCI vendor ID
			info := &GPUInfo{
				Vendor:          "Intel",
				CodecsSupported: []string{"H.264", "HEVC"},
				MaxResolution:   "1080p",
				MaxRefreshHz:    60,
				NVENCSessions:   0,
				ThermalHeadroom: 0.15,
				Model:           "Intel GPU",
				VRAMMB:          4096,
			}
			return info
		}
	}
	return nil
}

func fallbackGPU() GPUInfo {
	// Check for Apple Silicon
	if _, err := os.Stat("/System/Library/CoreServices/SystemVersion.plist"); err == nil {
		return GPUInfo{
			Vendor:          "Apple",
			Model:           "Apple Silicon",
			CodecsSupported: []string{"H.264", "HEVC", "ProRes"},
			MaxResolution:   "4K",
			MaxRefreshHz:    120,
			NVENCSessions:   0,
			ThermalHeadroom: 0.2,
			VRAMMB:          16384,
		}
	}

	return GPUInfo{
		Vendor:          "Unknown",
		Model:           "Unknown GPU",
		CodecsSupported: []string{"H.264"},
		MaxResolution:   "1080p",
		MaxRefreshHz:    60,
		NVENCSessions:   0,
		ThermalHeadroom: 0.1,
		VRAMMB:          4096,
	}
}

// hasNVIDIA reports whether an NVIDIA GPU is detected.
func hasNVIDIA() bool {
	return fileExists("/proc/driver/nvidia/gpus")
}

// hasAMD reports whether an AMD GPU is detected.
func hasAMD() bool {
	return fileExists("/sys/class/drm/card0/device/vendor") && fileContains("/sys/class/drm/card0/device/vendor", "0x1002")
}

// hasIntel reports whether an Intel GPU is detected.
func hasIntel() bool {
	return fileExists("/sys/class/drm/card0/device/vendor") && fileContains("/sys/class/drm/card0/device/vendor", "0x8086")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fileContains(path, substr string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), substr)
}

func gpuInfoPath(card int, filename string) string {
	return filepath.Join(fmt.Sprintf("/sys/class/drm/card%d/device", card), filename)
}

// CountNVIDIAGPUs returns the number of NVIDIA GPUs detected.
func CountNVIDIAGPUs() int {
	entries, err := os.ReadDir("/proc/driver/nvidia/gpus")
	if err != nil {
		return 0
	}
	return len(entries)
}
