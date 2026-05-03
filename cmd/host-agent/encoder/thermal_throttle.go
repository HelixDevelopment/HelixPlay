// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

// Package encoder provides video encoding backends.
// This file implements thermal-aware quality throttling to prevent
// GPU overheating during extended streaming sessions.
//
// T281: thermal-aware quality throttling.
package encoder

import ()

// ThermalThrottle monitors GPU temperature and reduces encoder
// quality/preset when thermal limits are approached.
type ThermalThrottle struct {
	maxTemp     float64 // °C
	throttleTemp float64
	active      bool
}

// NewThermalThrottle creates a thermal monitor with the given limits.
func NewThermalThrottle(throttleTemp, maxTemp float64) *ThermalThrottle {
	return &ThermalThrottle{
		maxTemp:      maxTemp,
		throttleTemp: throttleTemp,
		active:       true,
	}
}

// Check reads the current GPU temperature and returns a recommended
// encoder preset adjustment.
//   0 = no change, -1 = reduce quality, -2 = emergency throttle
func (t *ThermalThrottle) Check(currentTemp float64) int {
	if currentTemp >= t.maxTemp {
		return -2
	}
	if currentTemp >= t.throttleTemp {
		return -1
	}
	return 0
}

// ApplyPreset returns the encoder preset name adjusted for thermal state.
func (t *ThermalThrottle) ApplyPreset(basePreset string, adjustment int) string {
	_ = t
	presets := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}
	baseIdx := -1
	for i, p := range presets {
		if p == basePreset {
			baseIdx = i
			break
		}
	}
	if baseIdx < 0 {
		return basePreset
	}
	idx := baseIdx + adjustment
	if idx < 0 {
		idx = 0
	}
	if idx >= len(presets) {
		idx = len(presets) - 1
	}
	return presets[idx]
}

// Status returns the current thermal throttle state.
func (t *ThermalThrottle) Status() map[string]interface{} {
	_ = t
	return map[string]interface{}{
		"active":        t.active,
		"throttle_temp": t.throttleTemp,
		"max_temp":      t.maxTemp,
	}
}
