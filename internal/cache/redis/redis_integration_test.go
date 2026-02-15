package redis

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

func TestRedisCache_SetGet(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("INTEGRATION_TESTS not set")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	cache, err := NewCacheRepository(cfg.GetRedisOpts())
	if err != nil {
		t.Fatalf("failed to init cache: %v", err)
	}
	defer cache.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	value := &models.URL{
		ID:          1,
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		CreatedAt:   time.Now(),
	}

	if err := cache.Set(ctx, "url:test", value, time.Minute); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	got, err := cache.Get(ctx, "url:test")
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}
	if got.OriginalURL != value.OriginalURL {
		t.Fatalf("expected %s, got %s", value.OriginalURL, got.OriginalURL)
	}
}

func TestRedisCache_NotFound(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("INTEGRATION_TESTS not set")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	cache, err := NewCacheRepository(cfg.GetRedisOpts())
	if err != nil {
		t.Fatalf("failed to init cache: %v", err)
	}
	defer cache.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = cache.Get(ctx, "url:missing")
	if err == nil || !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected error for missing key")
	}
}

func TestRedisCache_TTLExpires(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("INTEGRATION_TESTS not set")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	cache, err := NewCacheRepository(cfg.GetRedisOpts())
	if err != nil {
		t.Fatalf("failed to init cache: %v", err)
	}
	defer cache.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	value := &models.URL{
		ID:          2,
		ShortCode:   "ttl",
		OriginalURL: "https://example.com/ttl",
		CreatedAt:   time.Now(),
	}

	if err := cache.Set(ctx, "url:ttl", value, 50*time.Millisecond); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	time.Sleep(75 * time.Millisecond)

	_, err = cache.Get(ctx, "url:ttl")
	if err == nil || !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected cache miss after ttl")
	}
}
