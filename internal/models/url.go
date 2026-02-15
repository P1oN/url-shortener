package models

import "time"

type URL struct {
	ID          int        `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type CreateURLOptions struct {
	OriginalURL string        `json:"original_url"`
	CustomCode  string        `json:"custom_code,omitempty"`
	ExpiresIn   time.Duration `json:"expires_in,omitempty"`
}
