package models

import "time"

// CapabilitySide represents which side of the handshake advertised
type CapabilitySide string

const (
	CapabilitySideHost   CapabilitySide = "host"
	CapabilitySideClient CapabilitySide = "client"
)

// Capability represents codec/transport/GPU capability advertisement exchanged during session negotiation
type Capability struct {
	ID                 string    `json:"id" db:"id"`
	SessionID          string    `json:"session_id" db:"session_id"`
	Side               CapabilitySide `json:"side" db:"side"`
	CodecsOffered      []string  `json:"codecs_offered" db:"codecs_offered"`
	TransportsOffered  []string  `json:"transports_offered" db:"transports_offered"`
	ResolutionsOffered []string  `json:"resolutions_offered" db:"resolutions_offered"`
	SelectedCodec      *string   `json:"selected_codec" db:"selected_codec"`
	SelectedTransport  *string   `json:"selected_transport" db:"selected_transport"`
	SelectedResolution *string   `json:"selected_resolution" db:"selected_resolution"`
	NegotiatedAt       *time.Time `json:"negotiated_at" db:"negotiated_at"`
}
