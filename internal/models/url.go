package models

import (
	"errors"
	"net/url"
	"time"
)

type URL struct {
	ID          int        `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

func (u *URL) Validate() error {
	parsedURL, err := url.Parse(u.OriginalURL)
	if err != nil {
		return err
	}

	if parsedURL.Scheme == "" {
		return ErrInvalidURLScheme
	}

	if parsedURL.Host == "" {
		return ErrInvalidURLHost
	}

	return nil
}

var (
	ErrInvalidURLScheme = errors.New("URL must contain (http/https)")
	ErrInvalidURLHost   = errors.New("URL must contain a host")
)

type CreateURLOptions struct {
	OriginalURL string        `json:"original_url"`
	CustomCode  string        `json:"custom_code,omitempty"`
	ExpiresIn   time.Duration `json:"expires_in,omitempty"`
}
