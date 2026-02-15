package config

import (
	"strings"
	"testing"
	"time"
)

func TestFromEnvAndValidate(t *testing.T) {
	env := []string{
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_USER=user",
		"DB_PASSWORD=pass",
		"DB_NAME=db",
		"REDIS_HOST=localhost",
		"REDIS_PORT=6379",
		"REDIS_PASSWORD=redispass",
		"REDIS_DB_NAME=1",
		"BASE_URL=http://localhost:8080/",
		"MIGRATIONS_PATH=file:///root/migrations",
		"API_KEY=secret",
		"ENABLE_SWAGGER=true",
		"ADDRESS=:8080",
		"READ_TIMEOUT=10s",
		"WRITE_TIMEOUT=11s",
		"IDLE_TIMEOUT=12s",
		"GRACEFUL_SHUTDOWN_TIMEOUT=3s",
		"REQUEST_TIMEOUT=4s",
		"CACHE_TTL=30m",
		"DB_MAX_OPEN_CONNS=20",
		"DB_MAX_IDLE_CONNS=5",
		"DB_CONN_MAX_LIFETIME=1h",
		"DB_CONN_MAX_IDLE_TIME=2m",
	}

	cfgEnv := FromEnv(env)
	if err := cfgEnv.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if cfgEnv.BaseURL != "http://localhost:8080" {
		t.Fatalf("expected base url trim, got %s", cfgEnv.BaseURL)
	}
	if cfgEnv.RedisDBName != 1 {
		t.Fatalf("expected redis db 1, got %d", cfgEnv.RedisDBName)
	}
	if cfgEnv.ReadTimeout != 10*time.Second {
		t.Fatalf("expected read timeout 10s, got %s", cfgEnv.ReadTimeout)
	}
	if cfgEnv.CacheTTL != 30*time.Minute {
		t.Fatalf("expected cache ttl 30m, got %s", cfgEnv.CacheTTL)
	}

	cfg := cfgEnv.ToConfig()
	if cfg.Server.Address != ":8080" {
		t.Fatalf("expected address :8080, got %s", cfg.Server.Address)
	}
	if !cfg.EnableSwagger {
		t.Fatalf("expected swagger enabled")
	}
}

func TestValidateMissing(t *testing.T) {
	cfgEnv := FromEnv([]string{
		"DB_HOST=localhost",
		"DB_PORT=5432",
	})

	err := cfgEnv.Validate()
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "missing required env vars") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaults(t *testing.T) {
	cfgEnv := FromEnv([]string{
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_USER=user",
		"DB_PASSWORD=pass",
		"DB_NAME=db",
		"REDIS_HOST=localhost",
		"REDIS_PORT=6379",
		"REDIS_PASSWORD=redispass",
		"BASE_URL=http://localhost:8080",
		"MIGRATIONS_PATH=file:///root/migrations",
		"API_KEY=secret",
		"ADDRESS=:8080",
	})

	if cfgEnv.CacheTTL != time.Hour {
		t.Fatalf("expected default cache ttl 1h, got %s", cfgEnv.CacheTTL)
	}
	if cfgEnv.DBMaxOpenConns != 25 {
		t.Fatalf("expected default max open conns 25, got %d", cfgEnv.DBMaxOpenConns)
	}
	if cfgEnv.EnableSwagger {
		t.Fatalf("expected default swagger disabled")
	}
}
