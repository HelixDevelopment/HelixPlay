// Package capability provides host hardware detection and capability advertisement.
package capability

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Fingerprint generates a stable hardware identifier by hashing
// CPU model, motherboard serial, and GPU serial.
func Fingerprint() (string, error) {
	parts := []string{
		cpuInfo(),
		mbSerial(),
		gpuSerial(),
		runtime.GOOS,
	}

	h := sha256.New()
	for _, p := range parts {
		if p != "" {
			_, _ = h.Write([]byte(p))
		}
	}

	sum := h.Sum(nil)
	if len(sum) == 0 {
		return "", fmt.Errorf("failed to generate hardware fingerprint")
	}

	return hex.EncodeToString(sum)[:32], nil
}

// cpuInfo returns a best-effort CPU identifier string.
func cpuInfo() string {
	// On Linux, try /proc/cpuinfo model name
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "model name\t:") {
				return strings.TrimSpace(line[len("model name\t:"):])
			}
		}
	}
	// Fallback to GOARCH
	return runtime.GOARCH
}

// mbSerial returns a best-effort motherboard serial number.
func mbSerial() string {
	// On Linux, try DMI sysfs entries
	paths := []string{
		"/sys/class/dmi/id/board_serial",
		"/sys/class/dmi/id/product_serial",
	}
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			s := strings.TrimSpace(string(data))
			if s != "" && s != "To be filled by O.E.M." {
				return s
			}
		}
	}
	return ""
}

// gpuSerial returns a best-effort GPU identifier.
func gpuSerial() string {
	// Try to read from the first NVIDIA GPU
	if data, err := os.ReadFile("/proc/driver/nvidia/gpus/0/information"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "Model:") {
				return strings.TrimSpace(line[len("Model:"):])
			}
		}
	}
	// Try lspci-like paths (some systems expose this)
	if data, err := os.ReadFile("/sys/class/drm/card0/device/vendor"); err == nil {
		vendor := strings.TrimSpace(string(data))
		if data2, err := os.ReadFile("/sys/class/drm/card0/device/device"); err == nil {
			device := strings.TrimSpace(string(data2))
			return fmt.Sprintf("%s:%s", vendor, device)
		}
	}
	return ""
}
