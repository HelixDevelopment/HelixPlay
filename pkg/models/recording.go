package models

import "time"

// RecordingFormat represents supported recording container formats
type RecordingFormat string

const (
	RecordingFormatMKV  RecordingFormat = "mkv"
	RecordingFormatFMP4 RecordingFormat = "fmp4"
)

// CloudSyncStatus represents the background sync state
type CloudSyncStatus string

const (
	CloudSyncPending  CloudSyncStatus = "pending"
	CloudSyncSyncing  CloudSyncStatus = "syncing"
	CloudSyncComplete CloudSyncStatus = "complete"
	CloudSyncFailed   CloudSyncStatus = "failed"
)

// Recording represents a DVR capture of a streaming session
type Recording struct {
	ID             string          `json:"id" db:"id"`
	SessionID      string          `json:"session_id" db:"session_id"`
	UserID         string          `json:"user_id" db:"user_id"`
	Format         RecordingFormat `json:"format" db:"format"`
	Codec          string          `json:"codec" db:"codec"`
	QualityPreset  string          `json:"quality_preset" db:"quality_preset"`
	FileSizeGB     float64         `json:"file_size_gb" db:"file_size_gb"`
	DurationSec    int             `json:"duration_sec" db:"duration_sec"`
	LocalPath      string          `json:"local_path" db:"local_path"`
	CloudSyncStatus CloudSyncStatus `json:"cloud_sync_status" db:"cloud_sync_status"`
	CloudURL       *string         `json:"cloud_url" db:"cloud_url"`
	HDRMetadata    map[string]any  `json:"hdr_metadata" db:"hdr_metadata"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}
