package service

import (
	"context"
	"time"

	"url-shortener-go/internal/models"
)

type Repository interface {
	Create(ctx context.Context, url *models.URL) error
	GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (*models.URL, error)
	IncrementClickCount(ctx context.Context, urlID int) error
	DeleteExpiredURLs(ctx context.Context) error
	Close() error
}

type Cache interface {
	Set(ctx context.Context, key string, value *models.URL, expiration time.Duration) error
	Get(ctx context.Context, key string) (*models.URL, error)
	Delete(ctx context.Context, key string) error
	Close() error
}
