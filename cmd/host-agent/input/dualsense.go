package input

import (
    "fmt"
    "sync"
)

type DualSense struct {
    mu               sync.RWMutex
    connected         bool
    triggerResistance map[string]float64
    hapticLeft        float64
    hapticRight       float64
}

func NewDualSense() *DualSense {
    return &DualSense{
        triggerResistance: make(map[string]float64),
    }
}

func (d *DualSense) Connect() error {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.connected = true
    return nil
}

func (d *DualSense) Disconnect() {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.connected = false
}

func (d *DualSense) IsConnected() bool {
    d.mu.RLock()
    defer d.mu.RUnlock()
    return d.connected
}

func (d *DualSense) SetHaptics(left, right float64) error {
    d.mu.Lock()
    defer d.mu.Unlock()
    if !d.connected {
        return fmt.Errorf("not connected")
    }
    d.hapticLeft = left
    d.hapticRight = right
    return nil
}

func (d *DualSense) SetTriggerResistance(trigger string, resistance float64) error {
    d.mu.Lock()
    defer d.mu.Unlock()
    if !d.connected {
        return fmt.Errorf("not connected")
    }
    d.triggerResistance[trigger] = resistance
    return nil
}

func (d *DualSense) GetTriggerResistance(trigger string) float64 {
    d.mu.RLock()
    defer d.mu.RUnlock()
    return d.triggerResistance[trigger]
}

func (d *DualSense) ReadIMU() (gyro [3]float64, accel [3]float64) {
    d.mu.RLock()
    defer d.mu.RUnlock()
    if !d.connected {
        return
    }
    gyro = [3]float64{0.1, 0.2, 0.3}
    accel = [3]float64{9.8, 0.0, 0.1}
    return
}
