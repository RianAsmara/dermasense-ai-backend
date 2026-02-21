package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OpenAI   OpenAIConfig
	Google   GoogleConfig
	App      AppConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig holds JWT settings.
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// OpenAIConfig holds OpenAI API settings.
type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

// GoogleConfig holds Google OAuth settings.
type GoogleConfig struct {
	ClientID string
}

// AppConfig holds application-specific settings.
type AppConfig struct {
	DefaultFreeScans int
	MaxFileSize      int64 // in bytes
	UploadDir        string
}

// Load reads environment variables and returns a Config.
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "dermasense"),
			Password: getEnv("DB_PASSWORD", "dermasense"),
			DBName:   getEnv("DB_NAME", "dermasense"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "change-me-in-production"),
			Expiration: getDurationEnv("JWT_EXPIRATION", 72*time.Hour),
		},
		OpenAI: OpenAIConfig{
			APIKey:  getEnv("OPENAI_API_KEY", ""),
			BaseURL: getEnv("OPENAI_BASE_URL", "https://ai.sumopod.com/v1"),
			Model:   getEnv("OPENAI_MODEL", "gpt-4.1-mini"),
			Timeout: getDurationEnv("OPENAI_TIMEOUT", 60*time.Second),
		},
		Google: GoogleConfig{
			ClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		},
		App: AppConfig{
			DefaultFreeScans: getIntEnv("DEFAULT_FREE_SCANS", 3),
			MaxFileSize:      getInt64Env("MAX_FILE_SIZE", 10*1024*1024), // 10MB
			UploadDir:        getEnv("UPLOAD_DIR", "./uploads"),
		},
	}
}

// DatabaseURL returns the PostgreSQL connection string.
func (c *Config) DatabaseURL() string {
	return "postgres://" + c.Database.User + ":" + c.Database.Password +
		"@" + c.Database.Host + ":" + c.Database.Port +
		"/" + c.Database.DBName + "?sslmode=" + c.Database.SSLMode
}

// RedisAddr returns the Redis address.
func (c *Config) RedisAddr() string {
	return c.Redis.Host + ":" + c.Redis.Port
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getInt64Env(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
