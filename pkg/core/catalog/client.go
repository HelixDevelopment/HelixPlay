package catalog

import (
	"context"
	"fmt"

	catalogizerv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/catalog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client provides access to the HelixPlay catalog service.
type Client struct {
	conn   *grpc.ClientConn
	client catalogizerv1.CatalogServiceClient
}

// NewClient dials the catalog service at the given address.
func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial catalog: %w", err)
	}
	return &Client{
		conn:   conn,
		client: catalogizerv1.NewCatalogServiceClient(conn),
	}, nil
}

// Close tears down the gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// ListGames fetches a page of games for a tenant.
func (c *Client) ListGames(ctx context.Context, tenantID string, pageSize int32, pageToken string) (*catalogizerv1.ListGamesResponse, error) {
	return c.client.ListGames(ctx, &catalogizerv1.ListGamesRequest{
		TenantId:  tenantID,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
}

// GetGame fetches a single game.
func (c *Client) GetGame(ctx context.Context, tenantID, gameID string) (*catalogizerv1.Game, error) {
	resp, err := c.client.GetGame(ctx, &catalogizerv1.GetGameRequest{
		TenantId: tenantID,
		GameId:   gameID,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// SearchGames performs a search query.
func (c *Client) SearchGames(ctx context.Context, tenantID, query string, pageSize int32) (*catalogizerv1.SearchGamesResponse, error) {
	return c.client.SearchGames(ctx, &catalogizerv1.SearchGamesRequest{
		TenantId: tenantID,
		Query:    query,
		PageSize: pageSize,
	})
}
