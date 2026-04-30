package input

import (
    "sync"
)

type HotPlug struct {
    mu           sync.RWMutex
    controllers  map[string]map[string]string // controllerID -> capabilities
}

func NewHotPlug() *HotPlug {
    return &HotPlug{
        controllers: make(map[string]map[string]string),
    }
}

func (h *HotPlug) ControllerConnected(id string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if _, exists := h.controllers[id]; !exists {
        h.controllers[id] = map[string]string{
            "type":     "gamepad",
            "haptics": "basic",
            "triggers": "adaptive",
            "imu":      "gyro+accel",
        }
    }
}

func (h *HotPlug) ControllerDisconnected(id string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    delete(h.controllers, id)
}

func (h *HotPlug) IsConnected(id string) bool {
    h.mu.RLock()
    defer h.mu.RUnlock()
    _, exists := h.controllers[id]
    return exists
}

func (h *HotPlug) GetCapabilites(id string) map[string]string {
    h.mu.RLock()
    defer h.mu.RUnlock()
    caps, exists := h.controllers[id]
    if !exists {
        return nil
    }
    // Return copy
    result := make(map[string]string)
    for k, v := range caps {
        result[k] = v
    }
    return result
}

func (h *HotPlug) UpdateCapability(id, key, value string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if caps, exists := h.controllers[id]; exists {
        caps[key] = value
    }
}

func (h *HotPlug) ListConnected() []string {
    h.mu.RLock()
    defer h.mu.RUnlock()
    result := make([]string, 0, len(h.controllers))
    for id := range h.controllers {
        result = append(result, id)
    }
    return result
}
