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

type Env struct {
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
	Address        string

	ReadTimeout             time.Duration
	WriteTimeout            time.Duration
	IdleTimeout             time.Duration
	GracefulShutdownTimeout time.Duration
	RequestTimeout          time.Duration
	CacheTTL                time.Duration

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration
}

func Load() (*Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err = godotenv.Load(); err != nil {
			log.Printf("Failed to load .env: %v", err)
		}
	}

	env := FromEnv(os.Environ())
	if err := env.Validate(); err != nil {
		return nil, err
	}

	return env.ToConfig(), nil
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

func FromEnv(envVars []string) Env {
	envMap := toEnvMap(envVars)
	return Env{
		DBHost:     getRequiredString(envMap, "DB_HOST"),
		DBPort:     getRequiredString(envMap, "DB_PORT"),
		DBUser:     getRequiredString(envMap, "DB_USER"),
		DBPassword: getRequiredString(envMap, "DB_PASSWORD"),
		DBName:     getRequiredString(envMap, "DB_NAME"),

		RedisHost:     getRequiredString(envMap, "REDIS_HOST"),
		RedisPort:     getRequiredString(envMap, "REDIS_PORT"),
		RedisPassword: getRequiredString(envMap, "REDIS_PASSWORD"),
		RedisDBName:   getInt(envMap, "REDIS_DB_NAME", 0),

		BaseURL:        strings.TrimRight(getRequiredString(envMap, "BASE_URL"), "/"),
		MigrationsPath: getRequiredString(envMap, "MIGRATIONS_PATH"),
		APIKey:         getRequiredString(envMap, "API_KEY"),
		Address:        getRequiredString(envMap, "ADDRESS"),

		ReadTimeout:             getDuration(envMap, "READ_TIMEOUT", 15*time.Second),
		WriteTimeout:            getDuration(envMap, "WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:             getDuration(envMap, "IDLE_TIMEOUT", 60*time.Second),
		GracefulShutdownTimeout: getDuration(envMap, "GRACEFUL_SHUTDOWN_TIMEOUT", 5*time.Second),
		RequestTimeout:          getDuration(envMap, "REQUEST_TIMEOUT", 5*time.Second),
		CacheTTL:                getDuration(envMap, "CACHE_TTL", 1*time.Hour),

		DBMaxOpenConns:    getInt(envMap, "DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getInt(envMap, "DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetime: getDuration(envMap, "DB_CONN_MAX_LIFETIME", 30*time.Minute),
		DBConnMaxIdleTime: getDuration(envMap, "DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
	}
}

func (e Env) Validate() error {
	var missing []string
	required := []struct {
		name  string
		value string
	}{
		{"DB_HOST", e.DBHost},
		{"DB_PORT", e.DBPort},
		{"DB_USER", e.DBUser},
		{"DB_PASSWORD", e.DBPassword},
		{"DB_NAME", e.DBName},
		{"REDIS_HOST", e.RedisHost},
		{"REDIS_PORT", e.RedisPort},
		{"REDIS_PASSWORD", e.RedisPassword},
		{"BASE_URL", e.BaseURL},
		{"MIGRATIONS_PATH", e.MigrationsPath},
		{"API_KEY", e.APIKey},
		{"ADDRESS", e.Address},
	}

	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			missing = append(missing, item.name)
		}
	}

	if len(missing) > 0 {
		return errors.New("missing required env vars: " + strings.Join(missing, ", "))
	}

	return nil
}

func (e Env) ToConfig() *Config {
	return &Config{
		DBHost:     e.DBHost,
		DBPort:     e.DBPort,
		DBUser:     e.DBUser,
		DBPassword: e.DBPassword,
		DBName:     e.DBName,

		RedisHost:     e.RedisHost,
		RedisPort:     e.RedisPort,
		RedisPassword: e.RedisPassword,
		RedisDBName:   e.RedisDBName,

		BaseURL:        strings.TrimRight(e.BaseURL, "/"),
		MigrationsPath: e.MigrationsPath,
		APIKey:         e.APIKey,
		Server: ServerConfig{
			Address:                 e.Address,
			ReadTimeout:             e.ReadTimeout,
			WriteTimeout:            e.WriteTimeout,
			IdleTimeout:             e.IdleTimeout,
			GracefulShutdownTimeout: e.GracefulShutdownTimeout,
		},
		CacheTTL:          e.CacheTTL,
		RequestTimeout:    e.RequestTimeout,
		DBMaxOpenConns:    e.DBMaxOpenConns,
		DBMaxIdleConns:    e.DBMaxIdleConns,
		DBConnMaxLifetime: e.DBConnMaxLifetime,
		DBConnMaxIdleTime: e.DBConnMaxIdleTime,
	}
}

func getRequiredString(envMap map[string]string, key string) string {
	value := strings.TrimSpace(envMap[key])
	return value
}

func getInt(envMap map[string]string, key string, defaultValue int) int {
	valueStr := strings.TrimSpace(envMap[key])
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid value for %s: %s. Using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

func getDuration(envMap map[string]string, key string, defaultValue time.Duration) time.Duration {
	valueStr := strings.TrimSpace(envMap[key])
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("Invalid duration for %s: %s. Using default: %s", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

func toEnvMap(envVars []string) map[string]string {
	envMap := make(map[string]string, len(envVars))
	for _, entry := range envVars {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		envMap[key] = value
	}
	return envMap
}
