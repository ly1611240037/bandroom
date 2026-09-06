package main

import (
	"context"
	"log"

	"github.com/ly1611240037/bandroom/backend/internal/config"
	"github.com/ly1611240037/bandroom/backend/internal/db"
	"github.com/ly1611240037/bandroom/backend/internal/seed"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	database, err := db.Open(ctx, cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err := db.Migrate(ctx, database); err != nil {
		log.Fatal(err)
	}
	if err := seed.SeedDemo(ctx, database); err != nil {
		log.Fatal(err)
	}
	log.Printf("demo data ready in %s", cfg.Database)
}
