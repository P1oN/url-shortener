package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"url-shortener-go/internal/models"
	"url-shortener-go/internal/service"

	"github.com/gorilla/mux"
)

type stubRepo struct {
	created *models.URL
}

func (s *stubRepo) Create(_ context.Context, url *models.URL) error {
	s.created = url
	return nil
}

func (s *stubRepo) GetByShortCode(_ context.Context, _ string) (*models.URL, error) {
	return nil, service.ErrNotFound
}

func (s *stubRepo) GetByOriginalURL(_ context.Context, _ string) (*models.URL, error) {
	return nil, service.ErrNotFound
}

func (s *stubRepo) IncrementClickCount(_ context.Context, _ int) error {
	return nil
}

func (s *stubRepo) DeleteExpiredURLs(_ context.Context) error {
	return nil
}

func (s *stubRepo) Close() error {
	return nil
}

type stubCache struct{}

func (s *stubCache) Set(_ context.Context, _ string, _ *models.URL, _ time.Duration) error {
	return nil
}

func (s *stubCache) Get(_ context.Context, _ string) (*models.URL, error) {
	return nil, service.ErrNotFound
}

func (s *stubCache) Delete(_ context.Context, _ string) error {
	return nil
}

func (s *stubCache) Close() error {
	return nil
}

func TestCreateShortURLHandler_Success(t *testing.T) {
	repo := &stubRepo{}
	cache := &stubCache{}
	svc := service.New(repo, cache, "http://localhost:8080", time.Hour, 2*time.Second)
	handlers := NewHandlers(svc)

	payload := map[string]interface{}{
		"original_url": "https://example.com",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/shorten", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handlers.CreateShortURLHandler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if repo.created == nil {
		t.Fatalf("expected create to be called")
	}
}

func TestCreateShortURLHandler_InvalidJSON(t *testing.T) {
	repo := &stubRepo{}
	cache := &stubCache{}
	svc := service.New(repo, cache, "http://localhost:8080", time.Hour, 2*time.Second)
	handlers := NewHandlers(svc)

	req := httptest.NewRequest(http.MethodPost, "/v1/shorten", bytes.NewBufferString("{bad json"))
	rec := httptest.NewRecorder()

	handlers.CreateShortURLHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateShortURLHandler_MissingURL(t *testing.T) {
	repo := &stubRepo{}
	cache := &stubCache{}
	svc := service.New(repo, cache, "http://localhost:8080", time.Hour, 2*time.Second)
	handlers := NewHandlers(svc)

	body, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest(http.MethodPost, "/v1/shorten", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handlers.CreateShortURLHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateShortURLHandler_NegativeExpires(t *testing.T) {
	repo := &stubRepo{}
	cache := &stubCache{}
	svc := service.New(repo, cache, "http://localhost:8080", time.Hour, 2*time.Second)
	handlers := NewHandlers(svc)

	negative := int64(-10)
	body, _ := json.Marshal(map[string]interface{}{
		"original_url":       "https://example.com",
		"expires_in_seconds": negative,
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/shorten", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handlers.CreateShortURLHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetFullURLHandler_NotFound(t *testing.T) {
	repo := &stubRepo{}
	cache := &stubCache{}
	svc := service.New(repo, cache, "http://localhost:8080", time.Hour, 2*time.Second)
	handlers := NewHandlers(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/missing", nil)
	req = mux.SetURLVars(req, map[string]string{"code": "missing"})
	rec := httptest.NewRecorder()

	handlers.GetFullURLHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHealthHandler_OK(t *testing.T) {
	handlers := NewHandlers(service.New(&stubRepo{}, &stubCache{}, "http://localhost:8080", time.Hour, 2*time.Second))
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()

	handlers.HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
