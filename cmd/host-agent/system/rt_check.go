// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

// Package system provides host-level configuration and tuning helpers.
// This file detects PREEMPT_RT kernel parameters and recommends
// real-time tuning for minimal streaming latency.
//
// T275: PREEMPT_RT kernel parameter detection and recommendation.
package system

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// RTCheck holds the results of a real-time readiness assessment.
type RTCheck struct {
	KernelRT      bool
	CPUGovernor   string
	TickRate      int
	Recommendations []string
}

// CheckRT analyses the current system for real-time suitability.
func CheckRT() (*RTCheck, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("PREEMPT_RT check requires Linux")
	}

	check := &RTCheck{}

	// Detect PREEMPT_RT in kernel version.
	data, err := os.ReadFile("/proc/version")
	if err == nil {
		version := string(data)
		check.KernelRT = strings.Contains(version, "PREEMPT_RT")
	}

	// Detect CPU governor.
	governorData, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor")
	if err == nil {
		check.CPUGovernor = strings.TrimSpace(string(governorData))
	}

	// Build recommendations.
	if !check.KernelRT {
		check.Recommendations = append(check.Recommendations,
			"Install a PREEMPT_RT patched kernel for sub-millisecond scheduling")
	}
	if check.CPUGovernor != "performance" {
		check.Recommendations = append(check.Recommendations,
			"Set CPU governor to 'performance': echo performance | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor")
	}

	return check, nil
}

// ApplyRecommendations attempts to apply safe kernel tuning.
// Requires CAP_SYS_ADMIN (typically root).
func ApplyRecommendations(check *RTCheck) error {
	_ = check
	return fmt.Errorf("automatic RT tuning not yet implemented")
}

// Report returns a human-readable report of the RT check.
func (c *RTCheck) Report() string {
	_ = c
	var b strings.Builder
	fmt.Fprintf(&b, "Kernel PREEMPT_RT: %v\n", c.KernelRT)
	fmt.Fprintf(&b, "CPU Governor: %s\n", c.CPUGovernor)
	if len(c.Recommendations) > 0 {
		fmt.Fprintln(&b, "Recommendations:")
		for _, r := range c.Recommendations {
			fmt.Fprintf(&b, "  - %s\n", r)
		}
	} else {
		fmt.Fprintln(&b, "System is optimally configured for real-time streaming.")
	}
	return b.String()
}
