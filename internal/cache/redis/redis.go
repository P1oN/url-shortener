package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"url-shortener-go/internal/models"
	"url-shortener-go/internal/service"

	"github.com/go-redis/redis"
)

type CacheRepository struct {
	client *redis.Client
}

func NewCacheRepository(opts *redis.Options) (*CacheRepository, error) {
	client := redis.NewClient(opts)

	_, err := client.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}

	return &CacheRepository{
		client: client,
	}, nil
}

func (r *CacheRepository) Set(ctx context.Context, key string, value *models.URL, expiration time.Duration) error {
	urlJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(key, urlJSON, expiration).Err()
}

func (r *CacheRepository) Get(ctx context.Context, key string) (*models.URL, error) {
	urlJSON, err := r.client.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, service.ErrNotFound
		}
		return nil, err
	}

	var url models.URL
	if err := json.Unmarshal([]byte(urlJSON), &url); err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(key).Err()
}

func (r *CacheRepository) Close() error {
	return r.client.Close()
}
