// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

// Package video provides HDR and colour-space processing for the host agent.
// This file implements HDR10/HDR10+ dynamic metadata extraction and
// forwarding.
//
// T280: HDR10/HDR10+ dynamic metadata pipeline.
package video

import (
	"fmt"
)

// HDRPipeline extracts HDR metadata from captured frames and forwards
// it alongside the video stream for client tone-mapping.
type HDRPipeline struct {
	standard string // "hdr10" or "hdr10+"
	enabled  bool
}

// NewHDRPipeline creates an HDR metadata pipeline.
func NewHDRPipeline(standard string) (*HDRPipeline, error) {
	if standard != "hdr10" && standard != "hdr10+" {
		return nil, fmt.Errorf("unsupported HDR standard: %s", standard)
	}
	return &HDRPipeline{standard: standard, enabled: true}, nil
}

// ExtractMetadata reads HDR metadata from a captured frame buffer.
func (h *HDRPipeline) ExtractMetadata(frame []byte, width, height int) (map[string]interface{}, error) {
	_ = h
	_ = frame
	_ = width
	_ = height
	return nil, fmt.Errorf("HDR metadata extraction not yet implemented")
}

// EmbedMetadata injects HDR SEI messages into the encoded bitstream.
func (h *HDRPipeline) EmbedMetadata(bitstream []byte, metadata map[string]interface{}) ([]byte, error) {
	_ = h
	_ = bitstream
	_ = metadata
	return nil, fmt.Errorf("HDR metadata embedding not yet implemented")
}

// Close releases HDR pipeline resources.
func (h *HDRPipeline) Close() error {
	_ = h
	return nil
}

// Standard returns the HDR standard in use.
func (h *HDRPipeline) Standard() string {
	_ = h
	return h.standard
}
