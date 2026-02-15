package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"url-shortener-go/internal/models"
)

type mockRepo struct {
	createCalls         int
	getByOriginalCalls  int
	getByShortCodeCalls int
	urlByOriginal       *models.URL
	urlByShortCode      *models.URL
	createErr           error
}

func (m *mockRepo) Create(_ context.Context, _ *models.URL) error {
	m.createCalls++
	return m.createErr
}

func (m *mockRepo) GetByShortCode(_ context.Context, _ string) (*models.URL, error) {
	m.getByShortCodeCalls++
	if m.urlByShortCode == nil {
		return nil, ErrNotFound
	}
	return m.urlByShortCode, nil
}

func (m *mockRepo) GetByOriginalURL(_ context.Context, _ string) (*models.URL, error) {
	m.getByOriginalCalls++
	if m.urlByOriginal == nil {
		return nil, ErrNotFound
	}
	return m.urlByOriginal, nil
}

func (m *mockRepo) IncrementClickCount(_ context.Context, _ int) error {
	return nil
}

func (m *mockRepo) DeleteExpiredURLs(_ context.Context) error {
	return nil
}

type mockCache struct {
	getCalls int
	url      *models.URL
}

func (m *mockCache) Set(_ context.Context, _ string, _ *models.URL, _ time.Duration) error {
	return nil
}

func (m *mockCache) Get(_ context.Context, _ string) (*models.URL, error) {
	m.getCalls++
	if m.url == nil {
		return nil, ErrNotFound
	}
	return m.url, nil
}

func TestCreateShortURL_InvalidURL(t *testing.T) {
	repo := &mockRepo{}
	cache := &mockCache{}
	svc := New(repo, cache, "http://localhost:8080", time.Hour, 2*time.Second)

	_, err := svc.CreateShortURL(context.Background(), models.CreateURLOptions{
		OriginalURL: "example.com",
	})
	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}

func TestCreateShortURL_ReusesExisting(t *testing.T) {
	existing := &models.URL{
		ID:          1,
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
	}
	repo := &mockRepo{urlByOriginal: existing}
	cache := &mockCache{}
	svc := New(repo, cache, "http://localhost:8080", time.Hour, 2*time.Second)

	url, err := svc.CreateShortURL(context.Background(), models.CreateURLOptions{
		OriginalURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url.ShortCode != "abc123" {
		t.Fatalf("expected existing short code, got %s", url.ShortCode)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected no create call, got %d", repo.createCalls)
	}
}

func TestGetFullURL_CacheHit(t *testing.T) {
	cached := &models.URL{
		ID:          2,
		ShortCode:   "cached",
		OriginalURL: "https://example.com",
	}
	repo := &mockRepo{}
	cache := &mockCache{url: cached}
	svc := New(repo, cache, "http://localhost:8080", time.Hour, 2*time.Second)

	url, err := svc.GetFullURL(context.Background(), "cached")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url.OriginalURL != "https://example.com" {
		t.Fatalf("unexpected url: %s", url.OriginalURL)
	}
	if repo.getByShortCodeCalls != 0 {
		t.Fatalf("expected repo to be unused, got %d calls", repo.getByShortCodeCalls)
	}
	if cache.getCalls != 1 {
		t.Fatalf("expected cache get call, got %d", cache.getCalls)
	}
}

func TestCreateShortURL_ConflictOnCustomCode(t *testing.T) {
	repo := &mockRepo{createErr: ErrConflict}
	cache := &mockCache{}
	svc := New(repo, cache, "http://localhost:8080", time.Hour, 2*time.Second)

	_, err := svc.CreateShortURL(context.Background(), models.CreateURLOptions{
		OriginalURL: "https://example.com",
		CustomCode:  "taken",
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}
