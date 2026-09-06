package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ly1611240037/bandroom/backend/internal/auth"
	"github.com/ly1611240037/bandroom/backend/internal/config"
	"github.com/ly1611240037/bandroom/backend/internal/db"
	"github.com/ly1611240037/bandroom/backend/internal/membership"
	"github.com/ly1611240037/bandroom/backend/internal/rooms"
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
	authHandler := auth.NewHandler(auth.NewService(database, cfg.AppURL, auth.LogMailer{Logf: log.Printf}))
	authHandler.RegisterRoutes(mux)
	membership.NewHandler(membership.NewService(database), authHandler).RegisterRoutes(mux)
	rooms.NewHandler(rooms.NewService(database), authHandler).RegisterRoutes(mux)

	server := &http.Server{Addr: cfg.HTTPAddr, Handler: cors(cfg.AppURL, mux)}
	log.Printf("BandRoom backend listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func cors(appURL string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", appURL)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
