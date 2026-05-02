// Package repository provides in-memory stub implementations for core backend
// services until persistent storage is wired in.
package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HelixDevelopment/HelixPlay/pkg/models"
	coreRepo "github.com/HelixDevelopment/HelixPlay/pkg/repository"
)

// MemorySessionRepository is an in-memory implementation of SessionRepository.
type MemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*models.Session
}

// NewMemorySessionRepository creates a new in-memory session repository.
func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{
		sessions: make(map[string]*models.Session),
	}
}

func (r *MemorySessionRepository) Create(_ context.Context, s *models.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = s
	return nil
}

func (r *MemorySessionRepository) GetByID(_ context.Context, id string) (*models.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %s not found", id)
	}
	return s, nil
}

func (r *MemorySessionRepository) Update(_ context.Context, s *models.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = s
	return nil
}

func (r *MemorySessionRepository) UpdateStatus(_ context.Context, id string, status models.SessionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}
	s.Status = status
	return nil
}

func (r *MemorySessionRepository) UpdateLatency(_ context.Context, id string, p50, p99, p999 float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}
	s.LatencyP50Ms = &p50
	s.LatencyP99Ms = &p99
	s.LatencyP999Ms = &p999
	return nil
}

func (r *MemorySessionRepository) Terminate(_ context.Context, id, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}
	s.Status = models.SessionStatusEnded
	now := time.Now().UTC()
	s.EndedAt = &now
	return nil
}

func (r *MemorySessionRepository) ListByUser(_ context.Context, userID string, limit, offset int) ([]*models.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*models.Session
	for _, s := range r.sessions {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return paginate(out, limit, offset), nil
}

func (r *MemorySessionRepository) ListByHost(_ context.Context, hostID string, limit, offset int) ([]*models.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*models.Session
	for _, s := range r.sessions {
		if s.HostID == hostID {
			out = append(out, s)
		}
	}
	return paginate(out, limit, offset), nil
}

func (r *MemorySessionRepository) ListActive(_ context.Context, limit, offset int) ([]*models.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*models.Session
	for _, s := range r.sessions {
		if s.Status == models.SessionStatusConnecting || s.Status == models.SessionStatusNegotiating || s.Status == models.SessionStatusStreaming {
			out = append(out, s)
		}
	}
	return paginate(out, limit, offset), nil
}

// MemoryTenantRepository is an in-memory implementation of TenantRepository.
type MemoryTenantRepository struct {
	mu      sync.RWMutex
	tenants map[string]*models.Tenant
	slugs   map[string]string // slug -> id
}

// NewMemoryTenantRepository creates a new in-memory tenant repository.
func NewMemoryTenantRepository() *MemoryTenantRepository {
	return &MemoryTenantRepository{
		tenants: make(map[string]*models.Tenant),
		slugs:   make(map[string]string),
	}
}

func (r *MemoryTenantRepository) Create(_ context.Context, t *models.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[t.ID] = t
	r.slugs[t.Slug] = t.ID
	return nil
}

func (r *MemoryTenantRepository) GetByID(_ context.Context, id string) (*models.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tenants[id]
	if !ok {
		return nil, fmt.Errorf("tenant %s not found", id)
	}
	return t, nil
}

func (r *MemoryTenantRepository) GetBySlug(_ context.Context, slug string) (*models.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.slugs[slug]
	if !ok {
		return nil, fmt.Errorf("tenant with slug %s not found", slug)
	}
	return r.tenants[id], nil
}

func (r *MemoryTenantRepository) Update(_ context.Context, t *models.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[t.ID] = t
	return nil
}

func (r *MemoryTenantRepository) UpdateTheme(_ context.Context, id, primaryColor, secondaryColor, logoURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tenants[id]
	if !ok {
		return fmt.Errorf("tenant %s not found", id)
	}
	t.ThemePrimaryColor = primaryColor
	t.ThemeSecondaryColor = secondaryColor
	t.ThemeLogoURL = logoURL
	return nil
}

func (r *MemoryTenantRepository) UpdateCatalogFilter(_ context.Context, id string, filter map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tenants[id]
	if !ok {
		return fmt.Errorf("tenant %s not found", id)
	}
	t.CatalogFilter = filter
	return nil
}

func (r *MemoryTenantRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tenants[id]
	if !ok {
		return fmt.Errorf("tenant %s not found", id)
	}
	delete(r.tenants, id)
	delete(r.slugs, t.Slug)
	return nil
}

func (r *MemoryTenantRepository) List(_ context.Context, limit, offset int) ([]*models.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*models.Tenant
	for _, t := range r.tenants {
		out = append(out, t)
	}
	return paginate(out, limit, offset), nil
}

// MemoryHostRepository is an in-memory implementation of HostRepository.
type MemoryHostRepository struct {
	mu    sync.RWMutex
	hosts map[string]*models.Host
}

// NewMemoryHostRepository creates a new in-memory host repository.
func NewMemoryHostRepository() *MemoryHostRepository {
	return &MemoryHostRepository{hosts: make(map[string]*models.Host)}
}

func (r *MemoryHostRepository) Create(_ context.Context, h *models.Host) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hosts[h.ID] = h
	return nil
}

func (r *MemoryHostRepository) GetByID(_ context.Context, id string) (*models.Host, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hosts[id]
	if !ok {
		return nil, fmt.Errorf("host %s not found", id)
	}
	return h, nil
}

func (r *MemoryHostRepository) GetByHardwareID(_ context.Context, hwID string) (*models.Host, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, h := range r.hosts {
		if h.HardwareID == hwID {
			return h, nil
		}
	}
	return nil, fmt.Errorf("host with hardware id %s not found", hwID)
}

func (r *MemoryHostRepository) Update(_ context.Context, h *models.Host) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hosts[h.ID] = h
	return nil
}

func (r *MemoryHostRepository) UpdateStatus(_ context.Context, id string, status models.HostStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	h, ok := r.hosts[id]
	if !ok {
		return fmt.Errorf("host %s not found", id)
	}
	h.Status = status
	return nil
}

func (r *MemoryHostRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.hosts, id)
	return nil
}

func (r *MemoryHostRepository) ListByTenant(_ context.Context, tenantID string, limit, offset int) ([]*models.Host, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Tenant filtering not supported in memory stub; return all hosts
	var out []*models.Host
	for _, h := range r.hosts {
		out = append(out, h)
	}
	return paginate(out, limit, offset), nil
}

func (r *MemoryHostRepository) ListOnline(_ context.Context, limit, offset int) ([]*models.Host, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*models.Host
	for _, h := range r.hosts {
		if h.Status == models.HostStatusOnline || h.Status == models.HostStatusStreaming {
			out = append(out, h)
		}
	}
	return paginate(out, limit, offset), nil
}

func paginate[T any](items []T, limit, offset int) []T {
	if offset >= len(items) {
		return nil
	}
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || offset+limit > len(items) {
		limit = len(items) - offset
	}
	return items[offset : offset+limit]
}

// Compile-time interface checks.
var (
	_ coreRepo.SessionRepository = (*MemorySessionRepository)(nil)
	_ coreRepo.TenantRepository  = (*MemoryTenantRepository)(nil)
	_ coreRepo.HostRepository    = (*MemoryHostRepository)(nil)
)
