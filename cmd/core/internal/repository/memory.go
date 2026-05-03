// Package repository provides in-memory stub implementations for core backend
// services until persistent storage is wired in.
package repository

import (
	"context"
	"fmt"
	"strings"
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

// MemoryGameRepository is an in-memory implementation of GameRepository.
type MemoryGameRepository struct {
	mu    sync.RWMutex
	games map[string]*models.Game
}

// NewMemoryGameRepository creates a new in-memory game repository.
func NewMemoryGameRepository() *MemoryGameRepository {
	return &MemoryGameRepository{games: make(map[string]*models.Game)}
}

func (r *MemoryGameRepository) Create(_ context.Context, g *models.Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.games[g.ID] = g
	return nil
}

func (r *MemoryGameRepository) GetByID(_ context.Context, id string) (*models.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.games[id]
	if !ok {
		return nil, fmt.Errorf("game %s not found", id)
	}
	return g, nil
}

func (r *MemoryGameRepository) GetByStoreID(_ context.Context, storeType models.GameStoreType, storeID string) (*models.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, g := range r.games {
		if g.StoreType == storeType && g.StoreID == storeID {
			return g, nil
		}
	}
	return nil, fmt.Errorf("game with store type %s and id %s not found", storeType, storeID)
}

func (r *MemoryGameRepository) Update(_ context.Context, g *models.Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.games[g.ID] = g
	return nil
}

func (r *MemoryGameRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.games, id)
	return nil
}

func (r *MemoryGameRepository) ListByHost(_ context.Context, hostID string, limit, offset int) ([]*models.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Host filtering not supported in memory stub; return all games
	var out []*models.Game
	for _, g := range r.games {
		out = append(out, g)
	}
	return paginate(out, limit, offset), nil
}

func (r *MemoryGameRepository) ListByTenant(_ context.Context, tenantID string, limit, offset int) ([]*models.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Tenant filtering not supported in memory stub; return all games
	var out []*models.Game
	for _, g := range r.games {
		out = append(out, g)
	}
	return paginate(out, limit, offset), nil
}

func (r *MemoryGameRepository) Search(_ context.Context, tenantID, query string, limit, offset int) ([]*models.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*models.Game
	for _, g := range r.games {
		if query == "" || containsIgnoreCase(g.Title, query) {
			out = append(out, g)
		}
	}
	return paginate(out, limit, offset), nil
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && len(substr) > 0 &&
		strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// MemoryUserRepository is an in-memory implementation of UserRepository.
type MemoryUserRepository struct {
	mu     sync.RWMutex
	users  map[string]*models.User
	emails map[string]string // tenantID:email -> userID
	oauth  map[string]string // tenantID:provider:subject -> userID
}

// NewMemoryUserRepository creates a new in-memory user repository.
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users:  make(map[string]*models.User),
		emails: make(map[string]string),
		oauth:  make(map[string]string),
	}
}

func (r *MemoryUserRepository) Create(_ context.Context, u *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	if u.Email != "" {
		r.emails[u.TenantID+":"+u.Email] = u.ID
	}
	if u.OAuth2Provider != "" && u.OAuth2Subject != "" {
		r.oauth[u.TenantID+":"+u.OAuth2Provider+":"+u.OAuth2Subject] = u.ID
	}
	return nil
}

func (r *MemoryUserRepository) GetByID(_ context.Context, id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user %s not found", id)
	}
	return u, nil
}

func (r *MemoryUserRepository) GetByEmail(_ context.Context, tenantID, email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.emails[tenantID+":"+email]
	if !ok {
		return nil, fmt.Errorf("user with email %s not found in tenant %s", email, tenantID)
	}
	return r.users[id], nil
}

func (r *MemoryUserRepository) GetByOAuthSubject(_ context.Context, tenantID, provider, subject string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.oauth[tenantID+":"+provider+":"+subject]
	if !ok {
		return nil, fmt.Errorf("user with oauth provider %s and subject %s not found in tenant %s", provider, subject, tenantID)
	}
	return r.users[id], nil
}

func (r *MemoryUserRepository) Update(_ context.Context, u *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.users[u.ID]
	if !ok {
		return fmt.Errorf("user %s not found", u.ID)
	}
	// Update email index
	if old.Email != u.Email || old.TenantID != u.TenantID {
		delete(r.emails, old.TenantID+":"+old.Email)
		if u.Email != "" {
			r.emails[u.TenantID+":"+u.Email] = u.ID
		}
	}
	// Update oauth index
	if old.OAuth2Provider != u.OAuth2Provider || old.OAuth2Subject != u.OAuth2Subject || old.TenantID != u.TenantID {
		delete(r.oauth, old.TenantID+":"+old.OAuth2Provider+":"+old.OAuth2Subject)
		if u.OAuth2Provider != "" && u.OAuth2Subject != "" {
			r.oauth[u.TenantID+":"+u.OAuth2Provider+":"+u.OAuth2Subject] = u.ID
		}
	}
	r.users[u.ID] = u
	return nil
}

func (r *MemoryUserRepository) UpdateLastLogin(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return fmt.Errorf("user %s not found", id)
	}
	now := time.Now().UTC()
	u.LastLoginAt = &now
	return nil
}

func (r *MemoryUserRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return nil
	}
	delete(r.users, id)
	delete(r.emails, u.TenantID+":"+u.Email)
	delete(r.oauth, u.TenantID+":"+u.OAuth2Provider+":"+u.OAuth2Subject)
	return nil
}

func (r *MemoryUserRepository) ListByTenant(_ context.Context, tenantID string, limit, offset int) ([]*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*models.User
	for _, u := range r.users {
		if u.TenantID == tenantID {
			out = append(out, u)
		}
	}
	return paginate(out, limit, offset), nil
}

// Compile-time interface checks.
var (
	_ coreRepo.SessionRepository = (*MemorySessionRepository)(nil)
	_ coreRepo.TenantRepository  = (*MemoryTenantRepository)(nil)
	_ coreRepo.HostRepository    = (*MemoryHostRepository)(nil)
	_ coreRepo.GameRepository    = (*MemoryGameRepository)(nil)
	_ coreRepo.UserRepository    = (*MemoryUserRepository)(nil)
)
