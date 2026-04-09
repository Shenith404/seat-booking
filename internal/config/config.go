package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Postgres  PostgresConfig
	Redis     RedisConfig
	Hold      HoldConfig
	RateLimit RateLimitConfig
	MailTrapConfig MailTrapConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int32
	MinConns int32
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// HoldConfig holds seat hold configuration
type HoldConfig struct {
	IdleTTL        time.Duration // 2 minutes sliding window
	MaxSessionTime time.Duration // 10 minutes absolute max
	MaxToggleCount int           // 15 actions per session
}

// RateLimitConfig holds rate limiter configuration
type RateLimitConfig struct {
	HoldRequestsPerMinute int // 20 requests per IP per minute for /hold
	BucketTTL             time.Duration
}

type MailTrapConfig struct {
	Host       string
	Port       int
	User       string
	Pass       string
	SenderName string
	SenderAddr string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "admin-user"),
			Password: getEnv("POSTGRES_PASSWORD", "password"),
			DBName:   getEnv("POSTGRES_DB", "cinema_db"),
			SSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),
			MaxConns: int32(getEnvAsInt("POSTGRES_MAX_CONNS", 25)),
			MinConns: int32(getEnvAsInt("POSTGRES_MIN_CONNS", 5)),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", "password"),
			DB:       getEnvAsInt("REDIS_DB", 0),
			PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 10),
		},
		Hold: HoldConfig{
			IdleTTL:        getEnvAsDuration("HOLD_IDLE_TTL", 2*time.Minute),
			MaxSessionTime: getEnvAsDuration("HOLD_MAX_SESSION_TIME", 10*time.Minute),
			MaxToggleCount: getEnvAsInt("HOLD_MAX_TOGGLE_COUNT", 15),
		},
		RateLimit: RateLimitConfig{
			HoldRequestsPerMinute: getEnvAsInt("RATE_LIMIT_HOLD_PER_MINUTE", 20),
			BucketTTL:             getEnvAsDuration("RATE_LIMIT_BUCKET_TTL", 1*time.Minute),
		},

		MailTrapConfig: MailTrapConfig{
			Host:       getEnv("MAILTRAP_HOST", "smtp.mailtrap.io"),
			Port:       getEnvAsInt("MAILTRAP_PORT", 587),
			User:       getEnv("MAILTRAP_USER", ""),
			Pass:       getEnv("MAILTRAP_PASS", ""),
			SenderName: getEnv("MAILTRAP_SENDER_NAME", "Sarislabs Seat Booking"),
			SenderAddr: getEnv("MAILTRAP_SENDER_ADDR", "noreply@sarislabs.com"),
		},
	}
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
