package catalog

import (
	"context"
	"net"
	"testing"
	"time"

	catalogizerv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/catalog"
	"google.golang.org/grpc"
)

type mockCatalogServer struct {
	catalogizerv1.UnimplementedCatalogServiceServer
	games []*catalogizerv1.Game
}

func (m *mockCatalogServer) ListGames(ctx context.Context, req *catalogizerv1.ListGamesRequest) (*catalogizerv1.ListGamesResponse, error) {
	return &catalogizerv1.ListGamesResponse{
		Games:     m.games,
		TotalSize: int64(len(m.games)),
	}, nil
}

func TestClientListGames(t *testing.T) {
	srv := grpc.NewServer()
	mock := &mockCatalogServer{
		games: []*catalogizerv1.Game{
			{GameId: "g1", Title: "Game One"},
			{GameId: "g2", Title: "Game Two"},
		},
	}
	catalogizerv1.RegisterCatalogServiceServer(srv, mock)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	go srv.Serve(lis)
	defer srv.Stop()

	client, err := NewClient(lis.Addr().String())
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.ListGames(ctx, "tenant-1", 10, "")
	if err != nil {
		t.Fatalf("ListGames failed: %v", err)
	}
	if len(resp.Games) != 2 {
		t.Fatalf("Expected 2 games, got %d", len(resp.Games))
	}
	if resp.Games[0].Title != "Game One" {
		t.Errorf("Unexpected title: %s", resp.Games[0].Title)
	}
}
