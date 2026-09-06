package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ly1611240037/bandroom/backend/internal/config"
	"github.com/ly1611240037/bandroom/backend/internal/db"
)

func main() {
	cfg := config.Load()
	database, err := db.Open(context.Background(), cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err := db.Migrate(context.Background(), database); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "bandroom-server"})
	})

	server := &http.Server{Addr: cfg.HTTPAddr, Handler: mux}
	log.Printf("BandRoom backend listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
