package models

import "time"

// GameStoreType represents supported game stores
type GameStoreType string

const (
	GameStoreSteam      GameStoreType = "steam"
	GameStoreEpic       GameStoreType = "epic"
	GameStoreGOG        GameStoreType = "gog"
	GameStoreUbisoft    GameStoreType = "ubisoft"
	GameStoreBattlenet  GameStoreType = "battlenet"
	GameStoreOrigin     GameStoreType = "origin"
	GameStoreMicrosoft  GameStoreType = "microsoft"
	GameStoreStandalone GameStoreType = "standalone"
)

// Game represents a game entry in the catalog, normalized across all store integrations
type Game struct {
	ID               string        `json:"id" db:"id"`
	StoreID          string        `json:"store_id" db:"store_id"`
	StoreType        GameStoreType `json:"store_type" db:"store_type"`
	Title            string        `json:"title" db:"title"`
	Description      string        `json:"description" db:"description"`
	CoverArtURL      string        `json:"cover_art_url" db:"cover_art_url"`
	ScreenshotURLs   []string      `json:"screenshot_urls" db:"screenshot_urls"`
	VideoURLs        []string      `json:"video_urls" db:"video_urls"`
	Genres           []string      `json:"genres" db:"genres"`
	ReleaseDate      *time.Time    `json:"release_date" db:"release_date"`
	RatingESRB       string        `json:"rating_esrb" db:"rating_esrb"`
	MinGPU           *string       `json:"min_gpu" db:"min_gpu"`
	RecGPU           *string       `json:"rec_gpu" db:"rec_gpu"`
	HDRSupport       bool          `json:"hdr_support" db:"hdr_support"`
	AtmosSupport     bool          `json:"atmos_support" db:"atmos_support"`
	DualSenseSupport bool          `json:"dualsense_support" db:"dualsense_support"`
	Multiplayer      bool          `json:"multiplayer" db:"multiplayer"`
	InstalledPath    *string       `json:"installed_path" db:"installed_path"`
	ExePath          *string       `json:"exe_path" db:"exe_path"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at"`
}
