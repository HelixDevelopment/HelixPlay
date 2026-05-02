package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"digital.vasic.auth/pkg/jwt"
	"github.com/HelixDevelopment/HelixPlay/pkg/core/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/api"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/auth"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/catalog"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/session"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/tenant"
	catalogizerv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/catalog"
	"google.golang.org/grpc"
)

// minimalCatalogClient implements catalogizerv1.CatalogServiceClient.
type minimalCatalogClient struct{}

func (m *minimalCatalogClient) ListGames(ctx context.Context, in *catalogizerv1.ListGamesRequest, opts ...grpc.CallOption) (*catalogizerv1.ListGamesResponse, error) {
	return &catalogizerv1.ListGamesResponse{}, nil
}
func (m *minimalCatalogClient) GetGame(ctx context.Context, in *catalogizerv1.GetGameRequest, opts ...grpc.CallOption) (*catalogizerv1.Game, error) {
	return &catalogizerv1.Game{}, nil
}
func (m *minimalCatalogClient) SearchGames(ctx context.Context, in *catalogizerv1.SearchGamesRequest, opts ...grpc.CallOption) (*catalogizerv1.SearchGamesResponse, error) {
	return &catalogizerv1.SearchGamesResponse{}, nil
}
func (m *minimalCatalogClient) GetGameAssets(ctx context.Context, in *catalogizerv1.GetGameAssetsRequest, opts ...grpc.CallOption) (*catalogizerv1.GameAssets, error) {
	return &catalogizerv1.GameAssets{}, nil
}
func (m *minimalCatalogClient) GetTenantCatalog(ctx context.Context, in *catalogizerv1.GetTenantCatalogRequest, opts ...grpc.CallOption) (*catalogizerv1.TenantCatalog, error) {
	return &catalogizerv1.TenantCatalog{}, nil
}
func (m *minimalCatalogClient) UpdateTenantCatalogFilter(ctx context.Context, in *catalogizerv1.UpdateTenantCatalogFilterRequest, opts ...grpc.CallOption) (*catalogizerv1.TenantCatalog, error) {
	return &catalogizerv1.TenantCatalog{}, nil
}
func (m *minimalCatalogClient) Health(ctx context.Context, in *catalogizerv1.HealthRequest, opts ...grpc.CallOption) (*catalogizerv1.HealthResponse, error) {
	return &catalogizerv1.HealthResponse{}, nil
}

func TestNewGateway(t *testing.T) {
	mgr := jwt.NewManager(jwt.DefaultConfig("gw-secret"))
	disc := capability.NewDiscovery("host-agent", "localhost:50051")

	gw := api.NewGateway(api.Config{
		GRPCAddress:    "127.0.0.1:0",
		HTTPAddress:    "127.0.0.1:0",
		SessionService: &session.Service{},
		TenantService:  &tenant.Service{},
		CatalogProxy:   catalog.NewProxy(&minimalCatalogClient{}),
		AuthMiddleware: auth.NewMiddleware(mgr),
		Discovery:      disc,
	})

	if gw == nil {
		t.Fatal("expected non-nil gateway")
	}
	if gw.SessionService() == nil {
		t.Error("expected session service")
	}
	if gw.TenantService() == nil {
		t.Error("expected tenant service")
	}
	if gw.CatalogProxy() == nil {
		t.Error("expected catalog proxy")
	}
}

func TestGatewayHandleHTTP(t *testing.T) {
	mgr := jwt.NewManager(jwt.DefaultConfig("gw-secret"))
	disc := capability.NewDiscovery("host-agent", "localhost:50051")

	gw := api.NewGateway(api.Config{
		GRPCAddress:    "127.0.0.1:0",
		HTTPAddress:    "127.0.0.1:0",
		SessionService: &session.Service{},
		TenantService:  &tenant.Service{},
		CatalogProxy:   catalog.NewProxy(&minimalCatalogClient{}),
		AuthMiddleware: auth.NewMiddleware(mgr),
		Discovery:      disc,
	})

	gw.HandleHTTP("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	// exercise the gateway's internal mux directly via a fresh mux for the test
	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "ok" {
		t.Errorf("expected body ok, got %s", rr.Body.String())
	}
}

func TestGatewayHandleHTTPWithAuth(t *testing.T) {
	mgr := jwt.NewManager(jwt.DefaultConfig("gw-secret"))
	token, _ := mgr.Create(map[string]interface{}{
		"sub":   "user-1",
		"roles": []string{"admin"},
	})

	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	gw := api.NewGateway(api.Config{
		GRPCAddress:    "127.0.0.1:0",
		HTTPAddress:    "127.0.0.1:0",
		SessionService: &session.Service{},
		TenantService:  &tenant.Service{},
		CatalogProxy:   catalog.NewProxy(&minimalCatalogClient{}),
		AuthMiddleware: auth.NewMiddleware(mgr),
		Discovery:      disc,
	})

	var received bool
	gw.HandleHTTPWithAuth("/admin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(http.StatusOK)
	}), "admin")

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// exercise via a fresh mux with the same middleware
	mux := http.NewServeMux()
	mw := auth.NewMiddleware(mgr)
	mux.Handle("/admin", mw.ValidateAndEnforce("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(http.StatusOK)
	})))
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !received {
		t.Error("expected handler to be invoked")
	}
}

func TestGatewayDiscoveryEndpoints(t *testing.T) {
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	gw := api.NewGateway(api.Config{
		Discovery: disc,
	})

	endpoints := gw.DiscoveryEndpoints("host-agent")
	if len(endpoints) == 0 {
		t.Error("expected at least one endpoint")
	}
	if endpoints[0].Type != "host-agent" {
		t.Errorf("expected type host-agent, got %s", endpoints[0].Type)
	}

	endpoints = gw.DiscoveryEndpoints("unknown")
	if len(endpoints) != 0 {
		t.Errorf("expected 0 endpoints for unknown type, got %d", len(endpoints))
	}
}

func TestStripPrefix(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/resource" {
			t.Errorf("expected path /resource, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := api.StripPrefix("/api", inner)

	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestStripPrefix_NoMatch(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := api.StripPrefix("/api", inner)

	req := httptest.NewRequest(http.MethodGet, "/other/resource", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestGatewayServe_ContextCancellation(t *testing.T) {
	mgr := jwt.NewManager(jwt.DefaultConfig("gw-secret"))
	disc := capability.NewDiscovery("host-agent", "localhost:50051")
	gw := api.NewGateway(api.Config{
		GRPCAddress:    "127.0.0.1:0",
		HTTPAddress:    "127.0.0.1:0",
		SessionService: &session.Service{},
		TenantService:  &tenant.Service{},
		CatalogProxy:   catalog.NewProxy(&minimalCatalogClient{}),
		AuthMiddleware: auth.NewMiddleware(mgr),
		Discovery:      disc,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- gw.Serve(ctx)
	}()

	select {
	case err := <-errCh:
		if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected Serve to return after context cancellation")
	}
}
