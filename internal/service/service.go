package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"url-shortener-go/internal/models"
	"url-shortener-go/pkg/utils"
)

type Service struct {
	repo           Repository
	cache          Cache
	baseURL        string
	cacheTTL       time.Duration
	requestTimeout time.Duration
}

func New(repo Repository, cache Cache, baseURL string, cacheTTL time.Duration, requestTimeout time.Duration) *Service {
	return &Service{
		repo:           repo,
		cache:          cache,
		baseURL:        baseURL,
		cacheTTL:       cacheTTL,
		requestTimeout: requestTimeout,
	}
}

func (s *Service) CreateShortURL(ctx context.Context, opts models.CreateURLOptions) (*models.URL, error) {
	if err := validateURL(opts.OriginalURL); err != nil {
		return nil, ErrInvalidURL
	}

	ctx, cancel := context.WithTimeout(ctx, s.requestTimeout)
	defer cancel()

	if opts.CustomCode == "" {
		existing, err := s.repo.GetByOriginalURL(ctx, opts.OriginalURL)
		if err == nil {
			return existing, nil
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}

	var expiresAt *time.Time
	if opts.ExpiresIn > 0 {
		expiresAt = new(time.Time(time.Now().Add(opts.ExpiresIn)))
	}

	shortCode := opts.CustomCode
	if shortCode == "" {
		for i := 0; i < 5; i++ {
			shortCode = utils.GenerateShortCode(6)
			url := &models.URL{
				ShortCode:   shortCode,
				OriginalURL: opts.OriginalURL,
				CreatedAt:   time.Now(),
				ExpiresAt:   expiresAt,
			}
			if err := s.repo.Create(ctx, url); err != nil {
				if errors.Is(err, ErrConflict) {
					continue
				}
				return nil, err
			}
			s.cache.Set(ctx, cacheKey(shortCode), url, s.cacheTTL)
			return url, nil
		}
		return nil, ErrConflict
	}

	newURL := &models.URL{
		ShortCode:   shortCode,
		OriginalURL: opts.OriginalURL,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
	}

	if err := s.repo.Create(ctx, newURL); err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey(shortCode), newURL, s.cacheTTL)

	return newURL, nil
}

func (s *Service) GetFullURL(ctx context.Context, shortCode string) (*models.URL, error) {
	ctx, cancel := context.WithTimeout(ctx, s.requestTimeout)
	defer cancel()

	if cached, err := s.cache.Get(ctx, cacheKey(shortCode)); err == nil {
		return cached, nil
	}

	url, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey(shortCode), url, s.cacheTTL)

	go func(urlID int) {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), s.requestTimeout)
		defer bgCancel()
		_ = s.repo.IncrementClickCount(bgCtx, urlID)
	}(url.ID)

	return url, nil
}

func (s *Service) GenerateShortURL(shortCode string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, shortCode)
}

func (s *Service) CleanupExpiredURLs(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.requestTimeout)
	defer cancel()
	return s.repo.DeleteExpiredURLs(ctx)
}

func validateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	if parsedURL.Scheme == "" {
		return ErrInvalidURL
	}

	if parsedURL.Host == "" {
		return ErrInvalidURL
	}

	return nil
}

func cacheKey(shortCode string) string {
	return "url:" + shortCode
}
