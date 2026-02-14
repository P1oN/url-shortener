package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"url-shortener-go/config"
	"url-shortener-go/internal/cache/redis"
	"url-shortener-go/internal/httpapi"
	httpmiddleware "url-shortener-go/internal/httpapi/middleware"
	"url-shortener-go/internal/repo/postgres"
	"url-shortener-go/internal/service"
	"url-shortener-go/internal/telemetry"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := telemetry.NewLogger()

	repo, err := postgres.NewRepository(cfg.GetPostgresConnString(), postgres.PoolConfig{
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
		ConnMaxIdleTime: cfg.DBConnMaxIdleTime,
	})
	if err != nil {
		log.Fatalf("Failed to create PostgreSQL repository: %v", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing repository: %v", err)
		}
	}()

	cache, err := redis.NewCacheRepository(cfg.GetRedisOpts())
	if err != nil {
		log.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer func() {
		if err := cache.Close(); err != nil {
			log.Printf("Error closing cache: %v", err)
		}
	}()

	service := service.New(repo, cache, cfg.BaseURL, cfg.CacheTTL, cfg.RequestTimeout)
	handlers := httpapi.NewHandlers(service)

	router := httpapi.SetupRoutes(handlers)

	router.Use(telemetry.RequestIDMiddleware)
	router.Use(telemetry.LoggingMiddleware(logger))
	router.Use(telemetry.RecoveryMiddleware(logger))
	router.Use(httpmiddleware.CorsMiddleware)
	router.Use(httpmiddleware.AuthMiddleware(cfg.APIKey))

	server := &http.Server{
		Handler:      router,
		Addr:         cfg.Server.Address,
		WriteTimeout: cfg.Server.WriteTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("server started", "address", cfg.Server.Address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("server exited gracefully")
}
