package catalog

import (
	"context"
	"fmt"

	catalogizerv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/catalog"
)

// Proxy wraps the Catalogizer gRPC client and provides a local API.
type Proxy struct {
	client catalogizerv1.CatalogServiceClient
}

// NewProxy creates a catalog proxy backed by the provided gRPC client.
func NewProxy(client catalogizerv1.CatalogServiceClient) *Proxy {
	return &Proxy{client: client}
}

// ListGames proxies a paginated game list request to the Catalogizer.
func (p *Proxy) ListGames(ctx context.Context, tenantID string, pageSize int32, pageToken string, filter *catalogizerv1.GameFilter, orderBy string) (*catalogizerv1.ListGamesResponse, error) {
	req := &catalogizerv1.ListGamesRequest{
		TenantId:  tenantID,
		PageSize:  pageSize,
		PageToken: pageToken,
		Filter:    filter,
		OrderBy:   orderBy,
	}
	resp, err := p.client.ListGames(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("catalog ListGames failed: %w", err)
	}
	return resp, nil
}

// GetGame proxies a single game lookup to the Catalogizer.
func (p *Proxy) GetGame(ctx context.Context, tenantID, gameID string, includeAssets bool) (*catalogizerv1.Game, error) {
	req := &catalogizerv1.GetGameRequest{
		TenantId:      tenantID,
		GameId:        gameID,
		IncludeAssets: includeAssets,
	}
	resp, err := p.client.GetGame(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("catalog GetGame failed: %w", err)
	}
	return resp, nil
}

// SearchGames proxies a search query to the Catalogizer.
func (p *Proxy) SearchGames(ctx context.Context, tenantID, query string, pageSize int32, pageToken string, facets []string) (*catalogizerv1.SearchGamesResponse, error) {
	req := &catalogizerv1.SearchGamesRequest{
		TenantId:  tenantID,
		Query:     query,
		PageSize:  pageSize,
		PageToken: pageToken,
		Facets:    facets,
	}
	resp, err := p.client.SearchGames(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("catalog SearchGames failed: %w", err)
	}
	return resp, nil
}

// GetGameAssets proxies an asset request to the Catalogizer.
func (p *Proxy) GetGameAssets(ctx context.Context, tenantID, gameID string, assetTypes []string, maxResolution, format string) (*catalogizerv1.GameAssets, error) {
	req := &catalogizerv1.GetGameAssetsRequest{
		TenantId:      tenantID,
		GameId:        gameID,
		AssetTypes:    assetTypes,
		MaxResolution: maxResolution,
		Format:        format,
	}
	resp, err := p.client.GetGameAssets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("catalog GetGameAssets failed: %w", err)
	}
	return resp, nil
}

// GetTenantCatalog proxies a tenant catalog request to the Catalogizer.
func (p *Proxy) GetTenantCatalog(ctx context.Context, tenantID string) (*catalogizerv1.TenantCatalog, error) {
	req := &catalogizerv1.GetTenantCatalogRequest{
		TenantId: tenantID,
	}
	resp, err := p.client.GetTenantCatalog(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("catalog GetTenantCatalog failed: %w", err)
	}
	return resp, nil
}

// UpdateTenantCatalogFilter proxies a filter update to the Catalogizer.
func (p *Proxy) UpdateTenantCatalogFilter(ctx context.Context, tenantID string, filter *catalogizerv1.CatalogFilter) (*catalogizerv1.TenantCatalog, error) {
	req := &catalogizerv1.UpdateTenantCatalogFilterRequest{
		TenantId: tenantID,
		Filter:   filter,
	}
	resp, err := p.client.UpdateTenantCatalogFilter(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("catalog UpdateTenantCatalogFilter failed: %w", err)
	}
	return resp, nil
}

// Health checks the upstream Catalogizer health.
func (p *Proxy) Health(ctx context.Context) (*catalogizerv1.HealthResponse, error) {
	req := &catalogizerv1.HealthRequest{}
	resp, err := p.client.Health(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("catalog Health failed: %w", err)
	}
	return resp, nil
}
