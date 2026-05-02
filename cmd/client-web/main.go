package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/client-web/handlers"
	"github.com/HelixDevelopment/HelixPlay/pkg/core/catalog"
)

func main() {
	mux := http.NewServeMux()

	streamingHandler := &handlers.Streaming{}
	streamingHandler.RegisterRoutes(mux)

	catalogAddr := os.Getenv("HELIXPLAY_CATALOG_ADDR")
	if catalogAddr != "" {
		catClient, err := catalog.NewClient(catalogAddr)
		if err != nil {
			log.Fatalf("Catalog client: %v", err)
		}
		catHandler := &handlers.Catalog{Client: catClient}
		catHandler.RegisterRoutes(mux)
	}

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Printf("Web client listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
}
