// Package repository provides comprehensive tests for in-memory repository implementations.
// Every test exercises real CRUD behavior to guarantee end-user usability (Constitution §1).
package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === Session Repository Tests ===

func TestMemorySessionRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := NewMemorySessionRepository()

	s := &models.Session{
		ID:     "sess-1",
		UserID: "user-1",
		HostID: "host-1",
		Status: models.SessionStatusConnecting,
		Codec:  "h264",
	}
	require.NoError(t, repo.Create(ctx, s))

	got, err := repo.GetByID(ctx, "sess-1")
	require.NoError(t, err)
	assert.Equal(t, "user-1", got.UserID)
	assert.Equal(t, models.SessionStatusConnecting, got.Status)
}

func TestMemorySessionRepository_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewMemorySessionRepository()

	_, err := repo.GetByID(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemorySessionRepository_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	repo := NewMemorySessionRepository()

	s := &models.Session{ID: "sess-1", Status: models.SessionStatusConnecting}
	require.NoError(t, repo.Create(ctx, s))

	require.NoError(t, repo.UpdateStatus(ctx, "sess-1", models.SessionStatusStreaming))

	got, err := repo.GetByID(ctx, "sess-1")
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusStreaming, got.Status)
}

func TestMemorySessionRepository_UpdateLatency(t *testing.T) {
	ctx := context.Background()
	repo := NewMemorySessionRepository()

	s := &models.Session{ID: "sess-1"}
	require.NoError(t, repo.Create(ctx, s))

	require.NoError(t, repo.UpdateLatency(ctx, "sess-1", 5.0, 12.0, 45.0))

	got, err := repo.GetByID(ctx, "sess-1")
	require.NoError(t, err)
	require.NotNil(t, got.LatencyP50Ms)
	require.NotNil(t, got.LatencyP99Ms)
	require.NotNil(t, got.LatencyP999Ms)
	assert.Equal(t, 5.0, *got.LatencyP50Ms)
	assert.Equal(t, 12.0, *got.LatencyP99Ms)
	assert.Equal(t, 45.0, *got.LatencyP999Ms)
}

func TestMemorySessionRepository_Terminate(t *testing.T) {
	ctx := context.Background()
	repo := NewMemorySessionRepository()

	s := &models.Session{ID: "sess-1", Status: models.SessionStatusStreaming}
	require.NoError(t, repo.Create(ctx, s))

	require.NoError(t, repo.Terminate(ctx, "sess-1", "user_disconnect"))

	got, err := repo.GetByID(ctx, "sess-1")
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusEnded, got.Status)
	assert.NotNil(t, got.EndedAt)
}

func TestMemorySessionRepository_ListByUser(t *testing.T) {
	ctx := context.Background()
	repo := NewMemorySessionRepository()

	require.NoError(t, repo.Create(ctx, &models.Session{ID: "s1", UserID: "u1"}))
	require.NoError(t, repo.Create(ctx, &models.Session{ID: "s2", UserID: "u1"}))
	require.NoError(t, repo.Create(ctx, &models.Session{ID: "s3", UserID: "u2"}))

	out, err := repo.ListByUser(ctx, "u1", 10, 0)
	require.NoError(t, err)
	assert.Len(t, out, 2)
}

func TestMemorySessionRepository_ListByHost(t *testing.T) {
	ctx := context.Background()
	repo := NewMemorySessionRepository()

	require.NoError(t, repo.Create(ctx, &models.Session{ID: "s1", HostID: "h1"}))
	require.NoError(t, repo.Create(ctx, &models.Session{ID: "s2", HostID: "h2"}))

	out, err := repo.ListByHost(ctx, "h1", 10, 0)
	require.NoError(t, err)
	assert.Len(t, out, 1)
	assert.Equal(t, "s1", out[0].ID)
}

func TestMemorySessionRepository_ListActive(t *testing.T) {
	ctx := context.Background()
	repo := NewMemorySessionRepository()

	require.NoError(t, repo.Create(ctx, &models.Session{ID: "s1", Status: models.SessionStatusStreaming}))
	require.NoError(t, repo.Create(ctx, &models.Session{ID: "s2", Status: models.SessionStatusEnded}))
	require.NoError(t, repo.Create(ctx, &models.Session{ID: "s3", Status: models.SessionStatusNegotiating}))

	out, err := repo.ListActive(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, out, 2)
}

func TestMemorySessionRepository_Pagination(t *testing.T) {
	ctx := context.Background()
	repo := NewMemorySessionRepository()
	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Create(ctx, &models.Session{ID: string(rune('a' + i)), UserID: "u1"}))
	}

	out, err := repo.ListByUser(ctx, "u1", 2, 1)
	require.NoError(t, err)
	assert.Len(t, out, 2)
}

func TestMemorySessionRepository_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	repo := NewMemorySessionRepository()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := string(rune('a' + n%26))
			_ = repo.Create(ctx, &models.Session{ID: id, UserID: "u1"})
			_, _ = repo.GetByID(ctx, id)
			_, _ = repo.ListActive(ctx, 10, 0)
		}(i)
	}
	wg.Wait()
}

// === Tenant Repository Tests ===

func TestMemoryTenantRepository_CreateAndGetByID(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryTenantRepository()

	tn := &models.Tenant{ID: "t1", Slug: "acme", Name: "Acme Corp"}
	require.NoError(t, repo.Create(ctx, tn))

	got, err := repo.GetByID(ctx, "t1")
	require.NoError(t, err)
	assert.Equal(t, "acme", got.Slug)
	assert.Equal(t, "Acme Corp", got.Name)
}

func TestMemoryTenantRepository_GetBySlug(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryTenantRepository()

	require.NoError(t, repo.Create(ctx, &models.Tenant{ID: "t1", Slug: "acme", Name: "Acme"}))

	got, err := repo.GetBySlug(ctx, "acme")
	require.NoError(t, err)
	assert.Equal(t, "t1", got.ID)
}

func TestMemoryTenantRepository_UpdateTheme(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryTenantRepository()

	require.NoError(t, repo.Create(ctx, &models.Tenant{ID: "t1", Slug: "acme"}))
	require.NoError(t, repo.UpdateTheme(ctx, "t1", "#ff0000", "#00ff00", "https://logo.png"))

	got, err := repo.GetByID(ctx, "t1")
	require.NoError(t, err)
	assert.Equal(t, "#ff0000", got.ThemePrimaryColor)
	assert.Equal(t, "#00ff00", got.ThemeSecondaryColor)
	assert.Equal(t, "https://logo.png", got.ThemeLogoURL)
}

func TestMemoryTenantRepository_UpdateCatalogFilter(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryTenantRepository()

	require.NoError(t, repo.Create(ctx, &models.Tenant{ID: "t1", Slug: "acme"}))
	filter := map[string]any{"genre": "action", "rating": "mature"}
	require.NoError(t, repo.UpdateCatalogFilter(ctx, "t1", filter))

	got, err := repo.GetByID(ctx, "t1")
	require.NoError(t, err)
	assert.Equal(t, filter, got.CatalogFilter)
}

func TestMemoryTenantRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryTenantRepository()

	require.NoError(t, repo.Create(ctx, &models.Tenant{ID: "t1", Slug: "acme"}))
	require.NoError(t, repo.Delete(ctx, "t1"))

	_, err := repo.GetByID(ctx, "t1")
	assert.Error(t, err)
}

func TestMemoryTenantRepository_List(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryTenantRepository()

	require.NoError(t, repo.Create(ctx, &models.Tenant{ID: "t1", Slug: "acme"}))
	require.NoError(t, repo.Create(ctx, &models.Tenant{ID: "t2", Slug: "globex"}))

	out, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, out, 2)
}

// === Host Repository Tests ===

func TestMemoryHostRepository_CreateAndGetByID(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryHostRepository()

	h := &models.Host{ID: "h1", Name: "Gaming-PC-1", HardwareID: "hw-123", Status: models.HostStatusOnline}
	require.NoError(t, repo.Create(ctx, h))

	got, err := repo.GetByID(ctx, "h1")
	require.NoError(t, err)
	assert.Equal(t, "Gaming-PC-1", got.Name)
	assert.Equal(t, models.HostStatusOnline, got.Status)
}

func TestMemoryHostRepository_GetByHardwareID(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryHostRepository()

	require.NoError(t, repo.Create(ctx, &models.Host{ID: "h1", HardwareID: "hw-abc"}))

	got, err := repo.GetByHardwareID(ctx, "hw-abc")
	require.NoError(t, err)
	assert.Equal(t, "h1", got.ID)
}

func TestMemoryHostRepository_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryHostRepository()

	require.NoError(t, repo.Create(ctx, &models.Host{ID: "h1", Status: models.HostStatusOffline}))
	require.NoError(t, repo.UpdateStatus(ctx, "h1", models.HostStatusStreaming))

	got, err := repo.GetByID(ctx, "h1")
	require.NoError(t, err)
	assert.Equal(t, models.HostStatusStreaming, got.Status)
}

func TestMemoryHostRepository_ListOnline(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryHostRepository()

	require.NoError(t, repo.Create(ctx, &models.Host{ID: "h1", Status: models.HostStatusOnline}))
	require.NoError(t, repo.Create(ctx, &models.Host{ID: "h2", Status: models.HostStatusOffline}))
	require.NoError(t, repo.Create(ctx, &models.Host{ID: "h3", Status: models.HostStatusStreaming}))

	out, err := repo.ListOnline(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, out, 2)
}

func TestMemoryHostRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryHostRepository()

	require.NoError(t, repo.Create(ctx, &models.Host{ID: "h1"}))
	require.NoError(t, repo.Delete(ctx, "h1"))

	_, err := repo.GetByID(ctx, "h1")
	assert.Error(t, err)
}

// === Game Repository Tests ===

func TestMemoryGameRepository_CreateAndGetByID(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGameRepository()

	g := &models.Game{ID: "g1", Title: "Cyberpunk 2077", StoreType: models.GameStoreSteam, StoreID: "steam-1091500"}
	require.NoError(t, repo.Create(ctx, g))

	got, err := repo.GetByID(ctx, "g1")
	require.NoError(t, err)
	assert.Equal(t, "Cyberpunk 2077", got.Title)
	assert.Equal(t, models.GameStoreSteam, got.StoreType)
}

func TestMemoryGameRepository_GetByStoreID(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGameRepository()

	require.NoError(t, repo.Create(ctx, &models.Game{ID: "g1", StoreType: models.GameStoreSteam, StoreID: "12345"}))

	got, err := repo.GetByStoreID(ctx, models.GameStoreSteam, "12345")
	require.NoError(t, err)
	assert.Equal(t, "g1", got.ID)
}

func TestMemoryGameRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGameRepository()

	require.NoError(t, repo.Create(ctx, &models.Game{ID: "g1", Title: "Old Title"}))
	require.NoError(t, repo.Update(ctx, &models.Game{ID: "g1", Title: "New Title"}))

	got, err := repo.GetByID(ctx, "g1")
	require.NoError(t, err)
	assert.Equal(t, "New Title", got.Title)
}

func TestMemoryGameRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGameRepository()

	require.NoError(t, repo.Create(ctx, &models.Game{ID: "g1"}))
	require.NoError(t, repo.Delete(ctx, "g1"))

	_, err := repo.GetByID(ctx, "g1")
	assert.Error(t, err)
}

func TestMemoryGameRepository_Search(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGameRepository()

	require.NoError(t, repo.Create(ctx, &models.Game{ID: "g1", Title: "Elden Ring"}))
	require.NoError(t, repo.Create(ctx, &models.Game{ID: "g2", Title: "Dark Souls"}))
	require.NoError(t, repo.Create(ctx, &models.Game{ID: "g3", Title: "Cyberpunk 2077"}))

	out, err := repo.Search(ctx, "", "souls", 10, 0)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "Dark Souls", out[0].Title)
}

func TestMemoryGameRepository_Search_EmptyQueryReturnsAll(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGameRepository()

	require.NoError(t, repo.Create(ctx, &models.Game{ID: "g1", Title: "Game A"}))
	require.NoError(t, repo.Create(ctx, &models.Game{ID: "g2", Title: "Game B"}))

	out, err := repo.Search(ctx, "", "", 10, 0)
	require.NoError(t, err)
	assert.Len(t, out, 2)
}

func TestMemoryGameRepository_ListByHost(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGameRepository()

	require.NoError(t, repo.Create(ctx, &models.Game{ID: "g1"}))
	require.NoError(t, repo.Create(ctx, &models.Game{ID: "g2"}))

	out, err := repo.ListByHost(ctx, "h1", 10, 0)
	require.NoError(t, err)
	assert.Len(t, out, 2) // Memory stub returns all games
}

// === User Repository Tests ===

func TestMemoryUserRepository_CreateAndGetByID(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryUserRepository()

	u := &models.User{ID: "u1", TenantID: "t1", Email: "alice@example.com", DisplayName: "Alice"}
	require.NoError(t, repo.Create(ctx, u))

	got, err := repo.GetByID(ctx, "u1")
	require.NoError(t, err)
	assert.Equal(t, "Alice", got.DisplayName)
	assert.Equal(t, "alice@example.com", got.Email)
}

func TestMemoryUserRepository_GetByEmail(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryUserRepository()

	require.NoError(t, repo.Create(ctx, &models.User{ID: "u1", TenantID: "t1", Email: "alice@example.com"}))

	got, err := repo.GetByEmail(ctx, "t1", "alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, "u1", got.ID)
}

func TestMemoryUserRepository_GetByOAuthSubject(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryUserRepository()

	require.NoError(t, repo.Create(ctx, &models.User{ID: "u1", TenantID: "t1", OAuth2Provider: "google", OAuth2Subject: "sub-123"}))

	got, err := repo.GetByOAuthSubject(ctx, "t1", "google", "sub-123")
	require.NoError(t, err)
	assert.Equal(t, "u1", got.ID)
}

func TestMemoryUserRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryUserRepository()

	require.NoError(t, repo.Create(ctx, &models.User{ID: "u1", TenantID: "t1", Email: "old@example.com", DisplayName: "Old"}))
	require.NoError(t, repo.Update(ctx, &models.User{ID: "u1", TenantID: "t1", Email: "new@example.com", DisplayName: "New"}))

	got, err := repo.GetByID(ctx, "u1")
	require.NoError(t, err)
	assert.Equal(t, "New", got.DisplayName)
	assert.Equal(t, "new@example.com", got.Email)

	// Verify old email index is gone
	_, err = repo.GetByEmail(ctx, "t1", "old@example.com")
	assert.Error(t, err)

	// Verify new email index works
	got2, err := repo.GetByEmail(ctx, "t1", "new@example.com")
	require.NoError(t, err)
	assert.Equal(t, "u1", got2.ID)
}

func TestMemoryUserRepository_UpdateLastLogin(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryUserRepository()

	require.NoError(t, repo.Create(ctx, &models.User{ID: "u1"}))
	require.NoError(t, repo.UpdateLastLogin(ctx, "u1"))

	got, err := repo.GetByID(ctx, "u1")
	require.NoError(t, err)
	assert.NotNil(t, got.LastLoginAt)
	assert.WithinDuration(t, time.Now().UTC(), *got.LastLoginAt, 5*time.Second)
}

func TestMemoryUserRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryUserRepository()

	require.NoError(t, repo.Create(ctx, &models.User{ID: "u1", TenantID: "t1", Email: "alice@example.com", OAuth2Provider: "google", OAuth2Subject: "sub-123"}))
	require.NoError(t, repo.Delete(ctx, "u1"))

	_, err := repo.GetByID(ctx, "u1")
	assert.Error(t, err)
	_, err = repo.GetByEmail(ctx, "t1", "alice@example.com")
	assert.Error(t, err)
	_, err = repo.GetByOAuthSubject(ctx, "t1", "google", "sub-123")
	assert.Error(t, err)
}

func TestMemoryUserRepository_ListByTenant(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryUserRepository()

	require.NoError(t, repo.Create(ctx, &models.User{ID: "u1", TenantID: "t1"}))
	require.NoError(t, repo.Create(ctx, &models.User{ID: "u2", TenantID: "t1"}))
	require.NoError(t, repo.Create(ctx, &models.User{ID: "u3", TenantID: "t2"}))

	out, err := repo.ListByTenant(ctx, "t1", 10, 0)
	require.NoError(t, err)
	assert.Len(t, out, 2)
}

func TestMemoryUserRepository_UpdateChangesOAuthIndex(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryUserRepository()

	require.NoError(t, repo.Create(ctx, &models.User{ID: "u1", TenantID: "t1", OAuth2Provider: "google", OAuth2Subject: "old-sub"}))
	require.NoError(t, repo.Update(ctx, &models.User{ID: "u1", TenantID: "t1", OAuth2Provider: "google", OAuth2Subject: "new-sub"}))

	_, err := repo.GetByOAuthSubject(ctx, "t1", "google", "old-sub")
	assert.Error(t, err)

	got, err := repo.GetByOAuthSubject(ctx, "t1", "google", "new-sub")
	require.NoError(t, err)
	assert.Equal(t, "u1", got.ID)
}

func TestMemoryUserRepository_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryUserRepository()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := string(rune('a' + n%26))
			_ = repo.Create(ctx, &models.User{ID: id, TenantID: "t1", Email: id + "@test.com"})
			_, _ = repo.GetByID(ctx, id)
			_, _ = repo.ListByTenant(ctx, "t1", 10, 0)
		}(i)
	}
	wg.Wait()
}
