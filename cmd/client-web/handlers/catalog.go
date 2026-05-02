package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/catalog"
)

// Catalog holds catalog-related HTTP handlers.
type Catalog struct {
	Client *catalog.Client
}

// RegisterRoutes registers catalog endpoints on the supplied mux.
func (c *Catalog) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/catalog/games", c.Games)
	mux.HandleFunc("/api/catalog/search", c.Search)
}

// Games handles listing catalog games.
func (c *Catalog) Games(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "missing tenant_id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := c.Client.ListGames(ctx, tenantID, 20, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Search handles catalog search.
func (c *Catalog) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query().Get("q")
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "missing tenant_id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resp, err := c.Client.SearchGames(ctx, tenantID, query, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
