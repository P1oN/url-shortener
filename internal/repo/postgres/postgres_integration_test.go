package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"url-shortener-go/config"
	"url-shortener-go/internal/models"
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

func TestPostgresRepository_CreateAndGet(t *testing.T) {
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

	shortCode := "testcode"
	originalURL := "https://example.com/test"

	if err := repo.Create(context.Background(), &models.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
	}); err != nil {
		t.Fatalf("failed to create url: %v", err)
	}

	t.Cleanup(func() {
		_, _ = repo.db.Exec("DELETE FROM url_stats WHERE url_id IN (SELECT id FROM urls WHERE short_code = $1)", shortCode)
		_, _ = repo.db.Exec("DELETE FROM urls WHERE short_code = $1", shortCode)
	})

	got, err := repo.GetByShortCode(context.Background(), shortCode)
	if err != nil {
		t.Fatalf("failed to get by short code: %v", err)
	}
	if got.OriginalURL != originalURL {
		t.Fatalf("expected %s, got %s", originalURL, got.OriginalURL)
	}

	byOriginal, err := repo.GetByOriginalURL(context.Background(), originalURL)
	if err != nil {
		t.Fatalf("failed to get by original url: %v", err)
	}
	if byOriginal.ShortCode != shortCode {
		t.Fatalf("expected %s, got %s", shortCode, byOriginal.ShortCode)
	}

	if err := repo.IncrementClickCount(context.Background(), got.ID); err != nil {
		t.Fatalf("failed to increment click count: %v", err)
	}
}
