package transport

import (
    "sync"
)

type Policies struct {
    mu                sync.RWMutex
    networkCondition  string
    sqPreset         string
    vrrEnabled       bool
    vrrRefreshRate   int
}

func NewPolicies() *Policies {
    return &Policies{
        networkCondition: "stable",
        sqPreset:        "medium",
        vrrEnabled:      false,
    }
}

func (p *Policies) CalculateBitrate(availableKbps int) int {
    p.mu.RLock()
    defer p.mu.RUnlock()
    // ABR: Adaptive Bitrate - scale based on available bandwidth
    // Reserve 20% overhead for FEC/control
    usable := float64(availableKbps) * 0.8
    // Clip to reasonable range
    if usable < 500 {
        return 300000 // 300 kbps minimum
    }
    if usable > 50000 {
        return 40000000 // 40 Mbps maximum
    }
    return int(usable * 1000) // Convert to bps
}

func (p *Policies) SetNetworkCondition(condition string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.networkCondition = condition
}

func (p *Policies) IsFECEnabled() bool {
    p.mu.RLock()
    defer p.mu.RUnlock()
    // FEC for unstable network
    return p.networkCondition == "unstable"
}

func (p *Policies) SetQuality(preset string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.sqPreset = preset
}

func (p *Policies) GetSQPreset() string {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return p.sqPreset
}

func (p *Policies) EnableVRR(refreshRateHz int) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.vrrEnabled = true
    p.vrrRefreshRate = refreshRateHz
}

func (p *Policies) IsVRREnabled() bool {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return p.vrrEnabled
}

func (p *Policies) PaceFrame(frameTimeMs float64) bool {
    p.mu.RLock()
    defer p.mu.RUnlock()
    if !p.vrrEnabled {
        return true // No pacing if VRR disabled
    }
    // With VRR enabled, accept any reasonable frame time
    // VRR can handle variable refresh rates
    return frameTimeMs > 0 && frameTimeMs < 1000.0
}
