package models

import "time"

// AssetType represents the category of a catalog asset
type AssetType string

const (
	AssetTypeCoverArt   AssetType = "cover_art"
	AssetTypeScreenshot AssetType = "screenshot"
	AssetTypeVideo      AssetType = "video_trailer"
	AssetTypeIcon       AssetType = "icon"
)

// AssetFormat represents supported asset file formats
type AssetFormat string

const (
	AssetFormatJPG  AssetFormat = "jpg"
	AssetFormatPNG  AssetFormat = "png"
	AssetFormatWebP AssetFormat = "webp"
	AssetFormatMP4  AssetFormat = "mp4"
	AssetFormatWebM AssetFormat = "webm"
)

// Asset represents a 4K media asset managed by the catalog system
type Asset struct {
	ID         string      `json:"id" db:"id"`
	GameID     string      `json:"game_id" db:"game_id"`
	AssetType  AssetType   `json:"asset_type" db:"asset_type"`
	CDNURL     string      `json:"cdn_url" db:"cdn_url"`
	CDNExpiry  *time.Time  `json:"cdn_expiry" db:"cdn_expiry"`
	WidthPx    int         `json:"width_px" db:"width_px"`
	HeightPx   int         `json:"height_px" db:"height_px"`
	FileSizeMB float64     `json:"file_size_mb" db:"file_size_mb"`
	Format     AssetFormat `json:"format" db:"format"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
}
