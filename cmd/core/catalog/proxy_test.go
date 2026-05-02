package catalog_test

import (
	"context"
	"errors"
	"testing"

	"github.com/HelixDevelopment/HelixPlay/cmd/core/catalog"
	catalogizerv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/catalog"
	"google.golang.org/grpc"
)

// mockCatalogClient implements catalogizerv1.CatalogServiceClient for testing.
type mockCatalogClient struct {
	listGamesFunc                 func(ctx context.Context, in *catalogizerv1.ListGamesRequest, opts ...grpc.CallOption) (*catalogizerv1.ListGamesResponse, error)
	getGameFunc                   func(ctx context.Context, in *catalogizerv1.GetGameRequest, opts ...grpc.CallOption) (*catalogizerv1.Game, error)
	searchGamesFunc               func(ctx context.Context, in *catalogizerv1.SearchGamesRequest, opts ...grpc.CallOption) (*catalogizerv1.SearchGamesResponse, error)
	getGameAssetsFunc             func(ctx context.Context, in *catalogizerv1.GetGameAssetsRequest, opts ...grpc.CallOption) (*catalogizerv1.GameAssets, error)
	getTenantCatalogFunc          func(ctx context.Context, in *catalogizerv1.GetTenantCatalogRequest, opts ...grpc.CallOption) (*catalogizerv1.TenantCatalog, error)
	updateTenantCatalogFilterFunc func(ctx context.Context, in *catalogizerv1.UpdateTenantCatalogFilterRequest, opts ...grpc.CallOption) (*catalogizerv1.TenantCatalog, error)
	healthFunc                    func(ctx context.Context, in *catalogizerv1.HealthRequest, opts ...grpc.CallOption) (*catalogizerv1.HealthResponse, error)
}

func (m *mockCatalogClient) ListGames(ctx context.Context, in *catalogizerv1.ListGamesRequest, opts ...grpc.CallOption) (*catalogizerv1.ListGamesResponse, error) {
	return m.listGamesFunc(ctx, in, opts...)
}
func (m *mockCatalogClient) GetGame(ctx context.Context, in *catalogizerv1.GetGameRequest, opts ...grpc.CallOption) (*catalogizerv1.Game, error) {
	return m.getGameFunc(ctx, in, opts...)
}
func (m *mockCatalogClient) SearchGames(ctx context.Context, in *catalogizerv1.SearchGamesRequest, opts ...grpc.CallOption) (*catalogizerv1.SearchGamesResponse, error) {
	return m.searchGamesFunc(ctx, in, opts...)
}
func (m *mockCatalogClient) GetGameAssets(ctx context.Context, in *catalogizerv1.GetGameAssetsRequest, opts ...grpc.CallOption) (*catalogizerv1.GameAssets, error) {
	return m.getGameAssetsFunc(ctx, in, opts...)
}
func (m *mockCatalogClient) GetTenantCatalog(ctx context.Context, in *catalogizerv1.GetTenantCatalogRequest, opts ...grpc.CallOption) (*catalogizerv1.TenantCatalog, error) {
	return m.getTenantCatalogFunc(ctx, in, opts...)
}
func (m *mockCatalogClient) UpdateTenantCatalogFilter(ctx context.Context, in *catalogizerv1.UpdateTenantCatalogFilterRequest, opts ...grpc.CallOption) (*catalogizerv1.TenantCatalog, error) {
	return m.updateTenantCatalogFilterFunc(ctx, in, opts...)
}
func (m *mockCatalogClient) Health(ctx context.Context, in *catalogizerv1.HealthRequest, opts ...grpc.CallOption) (*catalogizerv1.HealthResponse, error) {
	return m.healthFunc(ctx, in, opts...)
}

func TestProxyListGames(t *testing.T) {
	ctx := context.Background()
	client := &mockCatalogClient{
		listGamesFunc: func(ctx context.Context, in *catalogizerv1.ListGamesRequest, opts ...grpc.CallOption) (*catalogizerv1.ListGamesResponse, error) {
			if in.TenantId != "tenant-1" {
				return nil, errors.New("wrong tenant")
			}
			return &catalogizerv1.ListGamesResponse{TotalSize: 42}, nil
		},
	}
	p := catalog.NewProxy(client)

	resp, err := p.ListGames(ctx, "tenant-1", 10, "", nil, "")
	if err != nil {
		t.Fatalf("ListGames failed: %v", err)
	}
	if resp.TotalSize != 42 {
		t.Errorf("expected total_size 42, got %d", resp.TotalSize)
	}
}

func TestProxyGetGame(t *testing.T) {
	ctx := context.Background()
	client := &mockCatalogClient{
		getGameFunc: func(ctx context.Context, in *catalogizerv1.GetGameRequest, opts ...grpc.CallOption) (*catalogizerv1.Game, error) {
			return &catalogizerv1.Game{GameId: in.GameId, Title: "Test Game"}, nil
		},
	}
	p := catalog.NewProxy(client)

	game, err := p.GetGame(ctx, "tenant-1", "game-1", true)
	if err != nil {
		t.Fatalf("GetGame failed: %v", err)
	}
	if game.GameId != "game-1" {
		t.Errorf("expected game-1, got %s", game.GameId)
	}
	if game.Title != "Test Game" {
		t.Errorf("expected title Test Game, got %s", game.Title)
	}
}

func TestProxySearchGames(t *testing.T) {
	ctx := context.Background()
	client := &mockCatalogClient{
		searchGamesFunc: func(ctx context.Context, in *catalogizerv1.SearchGamesRequest, opts ...grpc.CallOption) (*catalogizerv1.SearchGamesResponse, error) {
			return &catalogizerv1.SearchGamesResponse{TotalSize: 5}, nil
		},
	}
	p := catalog.NewProxy(client)

	resp, err := p.SearchGames(ctx, "tenant-1", "rpg", 10, "", nil)
	if err != nil {
		t.Fatalf("SearchGames failed: %v", err)
	}
	if resp.TotalSize != 5 {
		t.Errorf("expected total_size 5, got %d", resp.TotalSize)
	}
}

func TestProxyGetGameAssets(t *testing.T) {
	ctx := context.Background()
	client := &mockCatalogClient{
		getGameAssetsFunc: func(ctx context.Context, in *catalogizerv1.GetGameAssetsRequest, opts ...grpc.CallOption) (*catalogizerv1.GameAssets, error) {
			return &catalogizerv1.GameAssets{GameId: in.GameId}, nil
		},
	}
	p := catalog.NewProxy(client)

	assets, err := p.GetGameAssets(ctx, "tenant-1", "game-1", []string{"cover"}, "1080p", "webp")
	if err != nil {
		t.Fatalf("GetGameAssets failed: %v", err)
	}
	if assets.GameId != "game-1" {
		t.Errorf("expected game-1, got %s", assets.GameId)
	}
}

func TestProxyGetTenantCatalog(t *testing.T) {
	ctx := context.Background()
	client := &mockCatalogClient{
		getTenantCatalogFunc: func(ctx context.Context, in *catalogizerv1.GetTenantCatalogRequest, opts ...grpc.CallOption) (*catalogizerv1.TenantCatalog, error) {
			return &catalogizerv1.TenantCatalog{TenantId: in.TenantId}, nil
		},
	}
	p := catalog.NewProxy(client)

	cat, err := p.GetTenantCatalog(ctx, "tenant-1")
	if err != nil {
		t.Fatalf("GetTenantCatalog failed: %v", err)
	}
	if cat.TenantId != "tenant-1" {
		t.Errorf("expected tenant-1, got %s", cat.TenantId)
	}
}

func TestProxyUpdateTenantCatalogFilter(t *testing.T) {
	ctx := context.Background()
	client := &mockCatalogClient{
		updateTenantCatalogFilterFunc: func(ctx context.Context, in *catalogizerv1.UpdateTenantCatalogFilterRequest, opts ...grpc.CallOption) (*catalogizerv1.TenantCatalog, error) {
			return &catalogizerv1.TenantCatalog{TenantId: in.TenantId}, nil
		},
	}
	p := catalog.NewProxy(client)

	filter := &catalogizerv1.CatalogFilter{AllowedGenres: []string{"action"}}
	cat, err := p.UpdateTenantCatalogFilter(ctx, "tenant-1", filter)
	if err != nil {
		t.Fatalf("UpdateTenantCatalogFilter failed: %v", err)
	}
	if cat.TenantId != "tenant-1" {
		t.Errorf("expected tenant-1, got %s", cat.TenantId)
	}
}

func TestProxyHealth(t *testing.T) {
	ctx := context.Background()
	client := &mockCatalogClient{
		healthFunc: func(ctx context.Context, in *catalogizerv1.HealthRequest, opts ...grpc.CallOption) (*catalogizerv1.HealthResponse, error) {
			return &catalogizerv1.HealthResponse{}, nil
		},
	}
	p := catalog.NewProxy(client)

	_, err := p.Health(ctx)
	if err != nil {
		t.Fatalf("Health failed: %v", err)
	}
}

func TestProxyErrorPropagation(t *testing.T) {
	ctx := context.Background()
	client := &mockCatalogClient{
		listGamesFunc: func(ctx context.Context, in *catalogizerv1.ListGamesRequest, opts ...grpc.CallOption) (*catalogizerv1.ListGamesResponse, error) {
			return nil, errors.New("upstream down")
		},
	}
	p := catalog.NewProxy(client)

	_, err := p.ListGames(ctx, "tenant-1", 10, "", nil, "")
	if err == nil {
		t.Fatal("expected error from upstream")
	}
}
