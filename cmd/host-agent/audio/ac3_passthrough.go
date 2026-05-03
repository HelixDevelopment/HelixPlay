// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

// Package audio provides audio encoding and forwarding for the host agent.
// This file implements AC3/EAC3 passthrough for bitstream-forwarding
// of Dolby Digital audio from the game to the client.
//
// T278: AC3/EAC3 passthrough.
package audio

import (
	"fmt"
)

// AC3Passthrough forwards raw AC3/EAC3 frames without re-encoding.
type AC3Passthrough struct {
	enabled bool
	format  string // "ac3" or "eac3"
}

// NewAC3Passthrough creates a passthrough handler.
func NewAC3Passthrough(format string) (*AC3Passthrough, error) {
	if format != "ac3" && format != "eac3" {
		return nil, fmt.Errorf("unsupported AC3 format: %s", format)
	}
	return &AC3Passthrough{enabled: true, format: format}, nil
}

// Forward sends an AC3/EAC3 frame to the client unchanged.
func (a *AC3Passthrough) Forward(frame []byte) error {
	_ = a
	_ = frame
	return fmt.Errorf("AC3 passthrough not yet implemented")
}

// Close stops the passthrough stream.
func (a *AC3Passthrough) Close() error {
	_ = a
	return nil
}

// Format returns the detected bitstream format.
func (a *AC3Passthrough) Format() string {
	_ = a
	return a.format
}
