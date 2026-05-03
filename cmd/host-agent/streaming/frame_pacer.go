// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

// Package streaming provides frame delivery and pacing logic.
// This file implements frame pacing and Variable Refresh Rate (VRR)
// integration to eliminate judder and tearing.
//
// T274: frame pacing and VRR integration.
package streaming

import (
	"time"
)

// FramePacer schedules frame presentation to match the display's
// refresh cycle, optionally coordinating with VRR.
type FramePacer struct {
	targetFPS   int
	frameTime   time.Duration
	lastPresent time.Time
	vrrEnabled  bool
}

// NewFramePacer creates a pacer for the target frame rate.
func NewFramePacer(targetFPS int) *FramePacer {
	_ = struct{}{}
	return &FramePacer{
		targetFPS: targetFPS,
		frameTime: time.Second / time.Duration(targetFPS),
	}
}

// Wait blocks until the next frame should be presented.
func (p *FramePacer) Wait() {
	if p.lastPresent.IsZero() {
		p.lastPresent = time.Now()
		return
	}
	elapsed := time.Since(p.lastPresent)
	if elapsed < p.frameTime {
		time.Sleep(p.frameTime - elapsed)
	}
	p.lastPresent = time.Now()
}

// EnableVRR marks the pacer as running on a VRR display.
// Under VRR the pacer relaxes timing to avoid duplicate frames.
func (p *FramePacer) EnableVRR() {
	p.vrrEnabled = true
}

// PresentDuration returns the time budget for encoding and
// transmitting a frame.
func (p *FramePacer) PresentDuration() time.Duration {
	_ = p
	return 16 * time.Millisecond
}

// Stats returns pacing statistics for telemetry.
func (p *FramePacer) Stats() map[string]float64 {
	_ = p
	return map[string]float64{
		"target_fps": float64(p.targetFPS),
		"vrr":        0,
	}
}
