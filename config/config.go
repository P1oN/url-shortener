package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Address                 string
	ReadTimeout             time.Duration
	WriteTimeout            time.Duration
	IdleTimeout             time.Duration
	GracefulShutdownTimeout time.Duration
}

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDBName   int

	BaseURL        string
	MigrationsPath string
	APIKey         string

	Server ServerConfig

	CacheTTL          time.Duration
	RequestTimeout    time.Duration
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		DBHost:     getRequiredString("DB_HOST"),
		DBPort:     getRequiredString("DB_PORT"),
		DBUser:     getRequiredString("DB_USER"),
		DBPassword: getRequiredString("DB_PASSWORD"),
		DBName:     getRequiredString("DB_NAME"),

		RedisHost:     getRequiredString("REDIS_HOST"),
		RedisPort:     getRequiredString("REDIS_PORT"),
		RedisPassword: getRequiredString("REDIS_PASSWORD"),
		RedisDBName:   getInt("REDIS_DB_NAME", 0),

		BaseURL:        strings.TrimRight(getRequiredString("BASE_URL"), "/"),
		MigrationsPath: getRequiredString("MIGRATIONS_PATH"),
		APIKey:         getRequiredString("API_KEY"),
		Server: ServerConfig{
			Address:                 getRequiredString("ADDRESS"),
			ReadTimeout:             getDuration("READ_TIMEOUT", 15*time.Second),
			WriteTimeout:            getDuration("WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:             getDuration("IDLE_TIMEOUT", 60*time.Second),
			GracefulShutdownTimeout: getDuration("GRACEFUL_SHUTDOWN_TIMEOUT", 5*time.Second),
		},
		CacheTTL:          getDuration("CACHE_TTL", 1*time.Hour),
		RequestTimeout:    getDuration("REQUEST_TIMEOUT", 5*time.Second),
		DBMaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		DBConnMaxIdleTime: getDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) GetPostgresConnString() string {
	return "host=" + c.DBHost + " port=" + c.DBPort + " user=" + c.DBUser + " password=" + c.DBPassword + " dbname=" + c.DBName + " sslmode=disable"
}

func (c *Config) GetRedisOpts() *redis.Options {
	return &redis.Options{
		Addr:     c.RedisHost + ":" + c.RedisPort,
		Password: c.RedisPassword,
		DB:       c.RedisDBName,
	}
}

func (c *Config) Validate() error {
	var missing []string
	required := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"REDIS_HOST",
		"REDIS_PORT",
		"REDIS_PASSWORD",
		"BASE_URL",
		"MIGRATIONS_PATH",
		"API_KEY",
		"ADDRESS",
	}

	for _, key := range required {
		if value, exists := os.LookupEnv(key); !exists || strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return errors.New("missing required env vars: " + strings.Join(missing, ", "))
	}

	return nil
}

func getString(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getRequiredString(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(value) == "" {
		return ""
	}
	return value
}

func getInt(key string, defaultValue int) int {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid value for %s: %s. Using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("Invalid duration for %s: %s. Using default: %s", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}
