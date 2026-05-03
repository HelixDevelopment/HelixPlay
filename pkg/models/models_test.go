// Package models provides comprehensive tests verifying all domain structs
// can be serialized, deserialized, and hold sensible data (Constitution §1).
package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsset_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	expiry := now.Add(24 * time.Hour)
	original := &Asset{
		ID:         "asset-1",
		GameID:     "game-1",
		AssetType:  AssetTypeScreenshot,
		Format:     AssetFormatPNG,
		CDNURL:     "https://cdn.example.com/screenshot.png",
		CDNExpiry:  &expiry,
		WidthPx:    3840,
		HeightPx:   2160,
		FileSizeMB: 2.5,
		CreatedAt:  now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Asset
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.GameID, decoded.GameID)
	assert.Equal(t, original.AssetType, decoded.AssetType)
	assert.Equal(t, original.Format, decoded.Format)
	assert.Equal(t, original.CDNURL, decoded.CDNURL)
	assert.Equal(t, *original.CDNExpiry, *decoded.CDNExpiry)
	assert.Equal(t, original.WidthPx, decoded.WidthPx)
	assert.Equal(t, original.HeightPx, decoded.HeightPx)
	assert.Equal(t, original.FileSizeMB, decoded.FileSizeMB)
	assert.Equal(t, original.CreatedAt, decoded.CreatedAt)
}

func TestCapability_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	codec := "h264"
	transport := "webrtc"
	resolution := "3840x2160"
	original := &Capability{
		ID:                 "cap-1",
		SessionID:          "sess-1",
		Side:               CapabilitySideHost,
		CodecsOffered:      []string{"h264", "hevc", "av1"},
		TransportsOffered:  []string{"webrtc", "quic"},
		ResolutionsOffered: []string{"1920x1080", "2560x1440", "3840x2160"},
		SelectedCodec:      &codec,
		SelectedTransport:  &transport,
		SelectedResolution: &resolution,
		NegotiatedAt:       &now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Capability
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Side, decoded.Side)
	assert.Equal(t, original.CodecsOffered, decoded.CodecsOffered)
	assert.Equal(t, *original.SelectedCodec, *decoded.SelectedCodec)
	assert.Equal(t, *original.SelectedTransport, *decoded.SelectedTransport)
	assert.Equal(t, *original.SelectedResolution, *decoded.SelectedResolution)
	assert.Equal(t, *original.NegotiatedAt, *decoded.NegotiatedAt)
}

func TestController_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	latency := 800
	original := &Controller{
		ID:                     "ctrl-1",
		SessionID:              "sess-1",
		DeviceType:             ControllerDeviceDualSense,
		ConnectionType:         ControllerConnBluetooth,
		HapticsEnabled:         true,
		AdaptiveTriggerEnabled: true,
		GyroEnabled:            true,
		AudioJackEnabled:       true,
		PollRateHz:             1000,
		LatencyUs:              &latency,
		ConnectedAt:            now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Controller
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.DeviceType, decoded.DeviceType)
	assert.Equal(t, original.ConnectionType, decoded.ConnectionType)
	assert.True(t, decoded.HapticsEnabled)
	assert.True(t, decoded.AdaptiveTriggerEnabled)
	assert.True(t, decoded.GyroEnabled)
	assert.True(t, decoded.AudioJackEnabled)
	assert.Equal(t, original.PollRateHz, decoded.PollRateHz)
	assert.Equal(t, *original.LatencyUs, *decoded.LatencyUs)
}

func TestGame_JSONRoundTrip(t *testing.T) {
	releaseDate := time.Date(2020, 12, 10, 0, 0, 0, 0, time.UTC)
	minGPU := "GTX 1060"
	recGPU := "RTX 3070"
	installedPath := "/games/cyberpunk"
	original := &Game{
		ID:               "game-1",
		StoreID:          "1091500",
		StoreType:        GameStoreSteam,
		Title:            "Cyberpunk 2077",
		Description:      "Open-world RPG",
		CoverArtURL:      "https://cdn.example.com/cover.jpg",
		ScreenshotURLs:   []string{"https://cdn.example.com/ss1.jpg"},
		VideoURLs:        []string{"https://cdn.example.com/trailer.mp4"},
		Genres:           []string{"RPG", "Action"},
		ReleaseDate:      &releaseDate,
		RatingESRB:       "M",
		MinGPU:           &minGPU,
		RecGPU:           &recGPU,
		HDRSupport:       true,
		AtmosSupport:     true,
		DualSenseSupport: true,
		Multiplayer:      false,
		InstalledPath:    &installedPath,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Game
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.Title, decoded.Title)
	assert.Equal(t, original.StoreType, decoded.StoreType)
	assert.Equal(t, original.Genres, decoded.Genres)
	assert.Equal(t, *original.ReleaseDate, *decoded.ReleaseDate)
	assert.Equal(t, original.RatingESRB, decoded.RatingESRB)
	assert.Equal(t, *original.MinGPU, *decoded.MinGPU)
	assert.Equal(t, *original.RecGPU, *decoded.RecGPU)
	assert.True(t, decoded.HDRSupport)
	assert.True(t, decoded.AtmosSupport)
	assert.True(t, decoded.DualSenseSupport)
}

func TestGPU_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	sessions := 2
	thermal := 65
	original := &GPU{
		ID:              "gpu-1",
		HostID:          "host-1",
		Vendor:          GPUVendorNVIDIA,
		Model:           "RTX 4090",
		DriverVersion:   "545.23.06",
		VRAMGB:          24,
		Encoder:         "nvenc",
		CodecsSupported: []string{"h264", "hevc", "av1"},
		MaxResolution:   "7680x4320",
		MaxRefreshHz:    240,
		NVENCSessions:   &sessions,
		ThermalC:        &thermal,
		SessionCount:    1,
		UpdatedAt:       now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded GPU
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.Vendor, decoded.Vendor)
	assert.Equal(t, original.Model, decoded.Model)
	assert.Equal(t, original.VRAMGB, decoded.VRAMGB)
	assert.Equal(t, *original.NVENCSessions, *decoded.NVENCSessions)
	assert.Equal(t, *original.ThermalC, *decoded.ThermalC)
	assert.Equal(t, original.UpdatedAt, decoded.UpdatedAt)
}

func TestHost_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	lanIP := "192.168.1.100"
	wanEndpoint := "host-1.helixplay.example.com:50051"
	original := &Host{
		ID:              "host-1",
		Name:            "Gaming-PC-1",
		HardwareID:      "hw-abc-123",
		OS:              "linux",
		OSVersion:       "6.8.0",
		LastSeenAt:      &now,
		Status:          HostStatusOnline,
		LanIP:           &lanIP,
		WanEndpoint:     &wanEndpoint,
		DiscoveryPort:   50051,
		ThermalThrottle: false,
		CreatedAt:       now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Host
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.Status, decoded.Status)
	assert.Equal(t, original.HardwareID, decoded.HardwareID)
	assert.Equal(t, *original.LanIP, *decoded.LanIP)
	assert.Equal(t, *original.WanEndpoint, *decoded.WanEndpoint)
	assert.Equal(t, original.DiscoveryPort, decoded.DiscoveryPort)
}

func TestRecording_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	original := &Recording{
		ID:              "rec-1",
		SessionID:       "sess-1",
		UserID:          "user-1",
		Format:          RecordingFormatMKV,
		Codec:           "hevc",
		QualityPreset:   "high",
		LocalPath:       "/recordings/rec-1.mkv",
		FileSizeGB:      4.5,
		DurationSec:     7200,
		CloudSyncStatus: CloudSyncComplete,
		CreatedAt:       now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Recording
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.Format, decoded.Format)
	assert.Equal(t, original.Codec, decoded.Codec)
	assert.Equal(t, original.LocalPath, decoded.LocalPath)
	assert.Equal(t, original.FileSizeGB, decoded.FileSizeGB)
	assert.Equal(t, original.DurationSec, decoded.DurationSec)
	assert.Equal(t, original.CloudSyncStatus, decoded.CloudSyncStatus)
}

func TestSession_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	gameID := "game-1"
	gpuID := "gpu-1"
	p50 := 12.5
	p99 := 45.0
	p999 := 120.0
	bandwidth := 50.0
	original := &Session{
		ID:              "sess-1",
		UserID:          "user-1",
		HostID:          "host-1",
		GameID:          &gameID,
		GPUID:           &gpuID,
		Status:          SessionStatusStreaming,
		Codec:           "hevc",
		Transport:       "webrtc",
		Resolution:      "3840x2160",
		RefreshHz:       120,
		LatencyP50Ms:    &p50,
		LatencyP99Ms:    &p99,
		LatencyP999Ms:   &p999,
		BandwidthMbps:   &bandwidth,
		ControllerType:  "dualsense",
		RecordingEnabled: true,
		StartedAt:       now,
		LastActivityAt:  now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Session
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.Status, decoded.Status)
	assert.Equal(t, original.Codec, decoded.Codec)
	assert.Equal(t, original.Resolution, decoded.Resolution)
	assert.Equal(t, *original.LatencyP50Ms, *decoded.LatencyP50Ms)
	assert.Equal(t, *original.LatencyP99Ms, *decoded.LatencyP99Ms)
	assert.Equal(t, *original.LatencyP999Ms, *decoded.LatencyP999Ms)
	assert.Equal(t, *original.BandwidthMbps, *decoded.BandwidthMbps)
	assert.Equal(t, original.ControllerType, decoded.ControllerType)
	assert.True(t, decoded.RecordingEnabled)
}

func TestTenant_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	original := &Tenant{
		ID:                       "t1",
		Slug:                     "acme-corp",
		Name:                     "Acme Corporation",
		ThemePrimaryColor:        "#ff0000",
		ThemeSecondaryColor:      "#00ff00",
		ThemeLogoURL:             "https://acme.example.com/logo.png",
		OAuth2Provider:           "google",
		OAuth2Config:             map[string]any{"client_id": "abc"},
		CatalogFilter:            map[string]any{"genre": "action"},
		MonetizationModel:        "subscription",
		ResourceQuotaMaxSessions: 100,
		ResourceQuotaStorageGB:   5000,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Tenant
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.Slug, decoded.Slug)
	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, original.ThemePrimaryColor, decoded.ThemePrimaryColor)
	assert.Equal(t, original.MonetizationModel, decoded.MonetizationModel)
	assert.Equal(t, original.ResourceQuotaMaxSessions, decoded.ResourceQuotaMaxSessions)
}

func TestUser_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	lastLogin := now.Add(-24 * time.Hour)
	original := &User{
		ID:                      "user-1",
		TenantID:                "t1",
		Email:                   "alice@example.com",
		DisplayName:             "Alice",
		Roles:                   []string{"player", "admin"},
		OAuth2Subject:           "sub-123",
		OAuth2Provider:          "google",
		PreferredControllerType: "dualsense",
		StorageQuotaUsedGB:      45.5,
		CreatedAt:               now,
		LastLoginAt:             &lastLogin,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded User
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.Email, decoded.Email)
	assert.Equal(t, original.DisplayName, decoded.DisplayName)
	assert.Equal(t, original.Roles, decoded.Roles)
	assert.Equal(t, original.OAuth2Provider, decoded.OAuth2Provider)
	assert.Equal(t, original.PreferredControllerType, decoded.PreferredControllerType)
	assert.Equal(t, original.StorageQuotaUsedGB, decoded.StorageQuotaUsedGB)
	assert.Equal(t, *original.LastLoginAt, *decoded.LastLoginAt)
}

func TestModelConstants(t *testing.T) {
	// Verify all exported constants have non-empty values
	assert.Equal(t, HostStatus("online"), HostStatusOnline)
	assert.Equal(t, SessionStatus("streaming"), SessionStatusStreaming)
	assert.Equal(t, GPUVendor("nvidia"), GPUVendorNVIDIA)
	assert.Equal(t, GameStoreType("steam"), GameStoreSteam)
	assert.Equal(t, ControllerDeviceType("dualsense"), ControllerDeviceDualSense)
	assert.Equal(t, AssetType("cover_art"), AssetTypeCoverArt)
	assert.Equal(t, RecordingFormat("mkv"), RecordingFormatMKV)
	assert.Equal(t, CloudSyncStatus("complete"), CloudSyncComplete)
	assert.Equal(t, CapabilitySide("host"), CapabilitySideHost)
}
