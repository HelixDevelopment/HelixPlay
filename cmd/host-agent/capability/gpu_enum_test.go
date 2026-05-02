package capability

import (
	"testing"
)

func TestDetectNVIDIA(t *testing.T) {
	// hasNVIDIA returns false in test environment without real GPU
	result := hasNVIDIA()
	// Result depends on test environment; just ensure it doesn't panic
	_ = result
}

func TestDetectAMD(t *testing.T) {
	result := hasAMD()
	_ = result
}

func TestDetectIntel(t *testing.T) {
	result := hasIntel()
	_ = result
}

func TestFallbackGPU(t *testing.T) {
	gpu := fallbackGPU()
	if gpu.Vendor == "" {
		t.Fatal("expected non-empty vendor")
	}
	if gpu.Model == "" {
		t.Fatal("expected non-empty model")
	}
	if len(gpu.CodecsSupported) == 0 {
		t.Fatal("expected at least one codec")
	}
	if gpu.MaxResolution == "" {
		t.Fatal("expected non-empty max resolution")
	}
	if gpu.MaxRefreshHz <= 0 {
		t.Fatal("expected positive refresh rate")
	}
}

func TestGPUInfoFields(t *testing.T) {
	gpu := GPUInfo{
		Vendor:          "NVIDIA",
		Model:           "RTX 4080",
		DriverVersion:   "545.23",
		VRAMMB:          16384,
		CodecsSupported: []string{"H.264", "HEVC", "AV1"},
		MaxResolution:   "4K",
		MaxRefreshHz:    240,
		NVENCSessions:   1,
		ThermalHeadroom: 0.3,
	}

	if gpu.Vendor != "NVIDIA" {
		t.Errorf("expected NVIDIA, got %s", gpu.Vendor)
	}
	if gpu.Model != "RTX 4080" {
		t.Errorf("expected RTX 4080, got %s", gpu.Model)
	}
	if gpu.VRAMMB != 16384 {
		t.Errorf("expected 16384 MB, got %d", gpu.VRAMMB)
	}
	if len(gpu.CodecsSupported) != 3 {
		t.Errorf("expected 3 codecs, got %d", len(gpu.CodecsSupported))
	}
}

func TestCountNVIDIAGPUs(t *testing.T) {
	count := CountNVIDIAGPUs()
	// In test environment without NVIDIA driver, should be 0
	if count < 0 {
		t.Fatal("expected non-negative GPU count")
	}
}
