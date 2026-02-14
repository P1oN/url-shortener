package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"url-shortener-go/config"
	"url-shortener-go/internal/service"
)

func TestPostgresRepository_GetByShortCode_NotFound(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("INTEGRATION_TESTS not set")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	repo, err := NewRepository(cfg.GetPostgresConnString(), PoolConfig{
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
		ConnMaxIdleTime: cfg.DBConnMaxIdleTime,
	})
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	defer repo.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = repo.GetByShortCode(ctx, "does-not-exist")
	if err == nil || !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
