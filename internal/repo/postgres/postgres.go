package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"url-shortener-go/internal/models"
	"url-shortener-go/internal/service"

	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func NewRepository(connStr string, pool PoolConfig) (*Repository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	db.SetMaxOpenConns(pool.MaxOpenConns)
	db.SetMaxIdleConns(pool.MaxIdleConns)
	db.SetConnMaxLifetime(pool.ConnMaxLifetime)
	db.SetConnMaxIdleTime(pool.ConnMaxIdleTime)

	if err = pingWithRetry(db, 10, 500*time.Millisecond); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return &Repository{db: db}, nil
}

func pingWithRetry(db *sql.DB, attempts int, baseDelay time.Duration) error {
	var err error
	delay := baseDelay
	for i := 0; i < attempts; i++ {
		err = db.Ping()
		if err == nil {
			return nil
		}
		time.Sleep(delay)
		if delay < 5*time.Second {
			delay *= 2
		}
	}
	return err
}

func (r *Repository) Create(ctx context.Context, url *models.URL) error {
	query := `
		WITH inserted_url AS (
			INSERT INTO urls (short_code, original_url, expires_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (short_code) DO NOTHING
			RETURNING id
		)
		INSERT INTO url_stats (url_id, click_count)
		SELECT id, 0
		FROM inserted_url
		WHERE id IS NOT NULL;
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, query,
		url.ShortCode,
		url.OriginalURL,
		url.ExpiresAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return service.ErrConflict
	}

	return tx.Commit()
}

func (r *Repository) GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	query := `
		SELECT id, original_url, created_at, expires_at
		FROM urls
		WHERE short_code = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`

	var url models.URL
	url.ShortCode = shortCode

	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.ExpiresAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *Repository) GetByOriginalURL(ctx context.Context, originalURL string) (*models.URL, error) {
	query := `
		SELECT id, short_code, created_at, expires_at
		FROM urls
		WHERE original_url = $1 AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
		LIMIT 1
	`

	var url models.URL
	url.OriginalURL = originalURL

	err := r.db.QueryRowContext(ctx, query, originalURL).Scan(
		&url.ID,
		&url.ShortCode,
		&url.CreatedAt,
		&url.ExpiresAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *Repository) IncrementClickCount(ctx context.Context, urlID int) error {
	query := `
		UPDATE url_stats
		SET
			click_count = click_count + 1,
			last_clicked_at = NOW()
		WHERE url_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, urlID)
	return err
}

func (r *Repository) DeleteExpiredURLs(ctx context.Context) error {
	query := `
		WITH deleted_urls AS (
			DELETE FROM urls
			WHERE expires_at < NOW()
			RETURNING id
		)
		DELETE FROM url_stats
		WHERE url_id IN (SELECT id FROM deleted_urls)
	`

	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *Repository) Close() error {
	return r.db.Close()
}
