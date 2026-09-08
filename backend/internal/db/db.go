package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Open(ctx context.Context, path string) (*sql.DB, error) {
	if path != ":memory:" && !strings.HasPrefix(path, "file:") {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}
	// DSN pragmas run for every pooled connection, not just the first one.
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	database, err := sql.Open("sqlite", path+separator+"_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_txlock=immediate")
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	if err := database.PingContext(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("connect sqlite database: %w", err)
	}
	return database, nil
}

func Migrate(ctx context.Context, database *sql.DB) error {
	if _, err := database.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		version, err := migrationVersion(entry.Name())
		if err != nil {
			return err
		}
		var count int
		if err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count); err != nil {
			return fmt.Errorf("check migration %d: %w", version, err)
		}
		if count > 0 {
			continue
		}
		content, err := migrationFiles.ReadFile(filepath.ToSlash(filepath.Join("migrations", entry.Name())))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", version, err)
		}
		if _, err := transaction.ExecContext(ctx, string(content)); err != nil {
			_ = transaction.Rollback()
			return fmt.Errorf("apply migration %d: %w", version, err)
		}
		if _, err := transaction.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES (?)", version); err != nil {
			_ = transaction.Rollback()
			return fmt.Errorf("record migration %d: %w", version, err)
		}
		if err := transaction.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", version, err)
		}
	}
	return nil
}

func migrationVersion(name string) (int, error) {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid migration filename: %s", name)
	}
	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid migration version in %s: %w", name, err)
	}
	return version, nil
}
