package capability

import (
	"encoding/json"
	"testing"
)

func TestFingerprintNotEmpty(t *testing.T) {
	fp, err := Fingerprint()
	if err != nil {
		t.Fatalf("Fingerprint failed: %v", err)
	}
	if len(fp) == 0 {
		t.Fatal("expected non-empty fingerprint")
	}
	if len(fp) != 32 {
		t.Fatalf("expected fingerprint length 32, got %d", len(fp))
	}
}

func TestFingerprintStable(t *testing.T) {
	fp1, err := Fingerprint()
	if err != nil {
		t.Fatalf("Fingerprint failed: %v", err)
	}
	fp2, err := Fingerprint()
	if err != nil {
		t.Fatalf("Fingerprint failed: %v", err)
	}
	if fp1 != fp2 {
		t.Fatalf("expected stable fingerprint, got %s vs %s", fp1, fp2)
	}
}

func TestEnumerateGPUs(t *testing.T) {
	gpus, err := EnumerateGPUs()
	if err != nil {
		t.Fatalf("EnumerateGPUs failed: %v", err)
	}
	if len(gpus) == 0 {
		t.Fatal("expected at least one GPU")
	}

	for i, gpu := range gpus {
		if gpu.Vendor == "" {
			t.Errorf("GPU %d: expected non-empty vendor", i)
		}
		if gpu.Model == "" {
			t.Errorf("GPU %d: expected non-empty model", i)
		}
		if len(gpu.CodecsSupported) == 0 {
			t.Errorf("GPU %d: expected at least one codec", i)
		}
		if gpu.MaxResolution == "" {
			t.Errorf("GPU %d: expected non-empty max resolution", i)
		}
		if gpu.MaxRefreshHz <= 0 {
			t.Errorf("GPU %d: expected positive refresh rate", i)
		}
	}
}

func TestAdvertise(t *testing.T) {
	meta, err := Advertise()
	if err != nil {
		t.Fatalf("Advertise failed: %v", err)
	}
	if meta.GPUModel == "" {
		t.Fatal("expected GPU model in advertisement")
	}
	if len(meta.CodecsSupported) == 0 {
		t.Fatal("expected at least one codec")
	}
	if meta.MaxResolution == "" {
		t.Fatal("expected max resolution")
	}
	if meta.RefreshRate <= 0 {
		t.Fatal("expected positive refresh rate")
	}
	if meta.GPUCount < 1 {
		t.Fatal("expected at least 1 GPU")
	}
}

func TestMetadataToJSON(t *testing.T) {
	meta := &Metadata{
		GPUModel:        "Test GPU",
		CodecsSupported: []string{"H.264"},
		MaxResolution:   "1080p",
		RefreshRate:     60,
	}

	jsonStr := meta.ToJSON()
	if jsonStr == "" {
		t.Fatal("expected non-empty JSON")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if parsed["gpu_model"] != "Test GPU" {
		t.Errorf("expected gpu_model Test GPU, got %v", parsed["gpu_model"])
	}
}
