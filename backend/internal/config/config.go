package config

import "os"

type Config struct {
	HTTPAddr     string
	Database     string
	AppURL       string
	MailFrom     string
	FrontendDist string
}

func Load() Config {
	return Config{
		HTTPAddr:     envOr("BANDROOM_HTTP_ADDR", ":8080"),
		Database:     envOr("BANDROOM_DATABASE", "./data/bandroom.db"),
		AppURL:       envOr("BANDROOM_APP_URL", "http://localhost:5173"),
		MailFrom:     envOr("BANDROOM_MAIL_FROM", "noreply@example.test"),
		FrontendDist: envOr("BANDROOM_FRONTEND_DIST", "./frontend/dist"),
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
