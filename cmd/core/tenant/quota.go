package tenant

import (
	"fmt"
	"sync"

	"github.com/HelixDevelopment/HelixPlay/pkg/models"
)

// Quota defines resource limits for a tenant.
type Quota struct {
	MaxSessions   int
	MaxStorageGB  int64
	MaxBandwidthMbps int
}

// DefaultQuota returns unrestricted defaults.
func DefaultQuota() Quota {
	return Quota{
		MaxSessions:      100,
		MaxStorageGB:     1000,
		MaxBandwidthMbps: 10000,
	}
}

// Enforcer tracks and enforces tenant resource quotas.
type Enforcer struct {
	mu       sync.RWMutex
	quotas   map[string]Quota
	usage    map[string]*Usage
}

// Usage tracks current resource consumption for a tenant.
type Usage struct {
	ActiveSessions int
	StorageGBUsed  int64
	BandwidthMbps  int
}

// NewEnforcer creates a quota enforcer.
func NewEnforcer() *Enforcer {
	return &Enforcer{
		quotas: make(map[string]Quota),
		usage:  make(map[string]*Usage),
	}
}

// SetQuota assigns a quota to a tenant.
func (e *Enforcer) SetQuota(tenantID string, q Quota) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.quotas[tenantID] = q
}

// GetQuota returns the quota for a tenant.
func (e *Enforcer) GetQuota(tenantID string) (Quota, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	q, ok := e.quotas[tenantID]
	return q, ok
}

// CanStartSession checks if a tenant can start a new session.
func (e *Enforcer) CanStartSession(tenantID string) error {
	e.mu.RLock()
	q, hasQuota := e.quotas[tenantID]
	u, hasUsage := e.usage[tenantID]
	e.mu.RUnlock()

	if !hasQuota {
		return fmt.Errorf("no quota defined for tenant %s", tenantID)
	}

	active := 0
	if hasUsage {
		active = u.ActiveSessions
	}

	if active >= q.MaxSessions {
		return fmt.Errorf("session quota exceeded: %d/%d", active, q.MaxSessions)
	}
	return nil
}

// RecordSessionStart increments the active session count.
func (e *Enforcer) RecordSessionStart(tenantID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.usage[tenantID] == nil {
		e.usage[tenantID] = &Usage{}
	}
	e.usage[tenantID].ActiveSessions++
}

// RecordSessionEnd decrements the active session count.
func (e *Enforcer) RecordSessionEnd(tenantID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.usage[tenantID] != nil && e.usage[tenantID].ActiveSessions > 0 {
		e.usage[tenantID].ActiveSessions--
	}
}

// Usage returns the current usage for a tenant.
func (e *Enforcer) Usage(tenantID string) Usage {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if u, ok := e.usage[tenantID]; ok {
		return *u
	}
	return Usage{}
}

// ApplyDefaults sets default quotas for tenants that don't have one.
func (e *Enforcer) ApplyDefaults(tenants []*models.Tenant) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, t := range tenants {
		if _, ok := e.quotas[t.ID]; !ok {
			e.quotas[t.ID] = Quota{
				MaxSessions:      t.ResourceQuotaMaxSessions,
				MaxStorageGB:     int64(t.ResourceQuotaStorageGB),
				MaxBandwidthMbps: 1000,
			}
		}
	}
}
