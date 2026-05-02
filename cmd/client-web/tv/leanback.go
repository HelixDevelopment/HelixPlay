package tv

import (
    "fmt"
    "sync"
)

type Leanback struct {
    mu         sync.RWMutex
    active     bool
    focusedItem string
    items      []string
    focusIndex  int
}

func NewLeanback() *Leanback {
    return &Leanback{
        items: []string{"Game 1", "Game 2", "Game 3", "Settings", "Quit"},
        focusIndex: 0,
    }
}

func (l *Leanback) Start() error {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.active = true
    l.focusedItem = l.items[0]
    return nil
}

func (l *Leanback) Stop() {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.active = false
}

func (l *Leanback) IsActive() bool {
    l.mu.RLock()
    defer l.mu.RUnlock()
    return l.active
}

func (l *Leanback) HandleDPad(direction string) error {
    l.mu.Lock()
    defer l.mu.Unlock()
    if !l.active {
        return fmt.Errorf("not active")
    }
    switch direction {
    case "Up", "Left":
        if l.focusIndex > 0 {
            l.focusIndex--
        }
    case "Down", "Right":
        if l.focusIndex < len(l.items)-1 {
            l.focusIndex++
        }
    case "Select":
    }
    l.focusedItem = l.items[l.focusIndex]
    return nil
}

func (l *Leanback) GetFocusedItem() string {
    l.mu.RLock()
    defer l.mu.RUnlock()
    return l.focusedItem
}

func (l *Leanback) MoveFocus(direction string) {
    l.HandleDPad(direction) //nolint
}
