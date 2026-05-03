// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

// Package audio provides audio encoding and forwarding for the host agent.
// This file implements Dolby Atmos metadata extraction and re-encode
// for object-based spatial audio forwarding.
//
// T279: Dolby Atmos forwarding (metadata extraction + re-encode).
package audio

import (
	"fmt"
)

// AtmosForwarder extracts Dolby Atmos metadata from the game audio
// stream and re-encodes it for client playback.
type AtmosForwarder struct {
	profile string
	enabled bool
}

// NewAtmosForwarder creates an Atmos forwarding pipeline.
func NewAtmosForwarder(profile string) (*AtmosForwarder, error) {
	if profile == "" {
		profile = "dd+"
	}
	return &AtmosForwarder{profile: profile, enabled: true}, nil
}

// ExtractMetadata parses the incoming audio bitstream for Atmos
// object metadata.
func (a *AtmosForwarder) ExtractMetadata(bitstream []byte) (map[string]interface{}, error) {
	_ = a
	_ = bitstream
	return nil, fmt.Errorf("Atmos metadata extraction not yet implemented")
}

// ReEncode re-encodes the audio with embedded Atmos metadata for
// the client.
func (a *AtmosForwarder) ReEncode(pcm []int16, metadata map[string]interface{}) ([]byte, error) {
	_ = a
	_ = pcm
	_ = metadata
	return nil, fmt.Errorf("Atmos re-encode not yet implemented")
}

// Close stops the Atmos forwarding pipeline.
func (a *AtmosForwarder) Close() error {
	_ = a
	return nil
}

// Profile returns the Atmos delivery profile (dd+, truehd, etc.).
func (a *AtmosForwarder) Profile() string {
	_ = a
	return a.profile
}
