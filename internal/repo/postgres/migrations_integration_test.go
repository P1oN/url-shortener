package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"url-shortener-go/config"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func TestMigrations_Up(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("INTEGRATION_TESTS not set")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.GetPostgresConnString())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}

	driver, err := migratepg.WithInstance(db, &migratepg.Config{})
	if err != nil {
		t.Fatalf("failed to create migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(cfg.MigrationsPath, "postgres", driver)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %v", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("failed to run migrations: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	assertRegclass(t, ctx, db, "public.urls")
	assertRegclass(t, ctx, db, "public.url_stats")
	assertRegclass(t, ctx, db, "public.idx_original_url")
}

func assertRegclass(t *testing.T, ctx context.Context, db *sql.DB, name string) {
	t.Helper()
	var regclass sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT to_regclass($1)", name).Scan(&regclass); err != nil {
		t.Fatalf("failed to query %s: %v", name, err)
	}
	if !regclass.Valid {
		t.Fatalf("expected %s to exist", name)
	}
}
