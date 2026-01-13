// Package config provides centralized configuration management.
// It loads configuration from environment variables with sensible defaults.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Server ServerConfig

	// Database configuration
	Database DatabaseConfig

	// CORS configuration
	CORS CORSConfig

	// RateLimit configuration
	RateLimit RateLimitConfig

	// Logging configuration
	Log LogConfig

	// JWT configuration
	JWT JWTConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	// Port to listen on
	Port string

	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration

	// IdleTimeout is the maximum duration an idle connection will remain open
	IdleTimeout time.Duration

	// ReadHeaderTimeout is the amount of time allowed to read request headers
	ReadHeaderTimeout time.Duration

	// ShutdownTimeout is the maximum duration to wait for active connections to close
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	// URL is the full database connection string (takes precedence if set)
	URL string

	// Host is the database server hostname
	Host string

	// Port is the database server port
	Port string

	// User is the database username
	User string

	// Password is the database password
	Password string

	// Name is the database name
	Name string

	// SSLMode is the SSL mode for the connection
	SSLMode string

	// MaxOpenConns is the maximum number of open connections
	MaxOpenConns int

	// MaxIdleConns is the maximum number of idle connections
	MaxIdleConns int

	// ConnMaxLifetime is the maximum lifetime of a connection
	ConnMaxLifetime time.Duration
}

// CORSConfig holds CORS middleware configuration
type CORSConfig struct {
	// AllowedOrigins is a comma-separated list of allowed origins
	AllowedOrigins []string

	// AllowedMethods is a comma-separated list of allowed HTTP methods
	AllowedMethods []string

	// AllowedHeaders is a comma-separated list of allowed headers
	AllowedHeaders []string

	// AllowCredentials indicates whether credentials are allowed
	AllowCredentials bool

	// MaxAge is the max age for preflight cache in seconds
	MaxAge int
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Enabled indicates whether rate limiting is enabled
	Enabled bool

	// Rate is the maximum requests per window
	Rate int

	// Window is the time window for rate limiting
	Window time.Duration
}

// LogConfig holds logging configuration
type LogConfig struct {
	// Level is the minimum log level (debug, info, warn, error)
	Level string

	// Format is the log format (json, text)
	Format string

	// AddSource indicates whether to include source file information
	AddSource bool
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	// SecretKey is the secret key used to sign JWT tokens
	SecretKey string

	// AccessTokenTTL is the access token time-to-live in minutes
	AccessTokenTTL int

	// RefreshTokenTTL is the refresh token time-to-live in hours
	RefreshTokenTTL int
}

// Load loads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:              getEnv("PORT", "8080"),
			ReadTimeout:       getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:      getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:       getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ReadHeaderTimeout: getDurationEnv("SERVER_READ_HEADER_TIMEOUT", 5*time.Second),
			ShutdownTimeout:   getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5433"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "app_db"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getSliceEnv("CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods:   getSliceEnv("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getSliceEnv("CORS_ALLOWED_HEADERS", []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"}),
			AllowCredentials: getBoolEnv("CORS_ALLOW_CREDENTIALS", false),
			MaxAge:           getIntEnv("CORS_MAX_AGE", 86400),
		},
		RateLimit: RateLimitConfig{
			Enabled: getBoolEnv("RATE_LIMIT_ENABLED", true),
			Rate:    getIntEnv("RATE_LIMIT_RATE", 100),
			Window:  getDurationEnv("RATE_LIMIT_WINDOW", time.Minute),
		},
		Log: LogConfig{
			Level:     getEnv("LOG_LEVEL", "info"),
			Format:    getEnv("LOG_FORMAT", "json"),
			AddSource: getBoolEnv("LOG_ADD_SOURCE", false),
		},
		JWT: JWTConfig{
			SecretKey:       getEnv("JWT_SECRET_KEY", "your-super-secret-key-change-in-production"),
			AccessTokenTTL:  getIntEnv("JWT_ACCESS_TOKEN_TTL", 15),  // 15 minutes
			RefreshTokenTTL: getIntEnv("JWT_REFRESH_TOKEN_TTL", 168), // 7 days (168 hours)
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv gets an integer environment variable or returns a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getBoolEnv gets a boolean environment variable or returns a default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// getDurationEnv gets a duration environment variable or returns a default value
// Accepts values like "30s", "5m", "1h"
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getSliceEnv gets a comma-separated environment variable as a slice
func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	env := getEnv("APP_ENV", "development")
	return env == "development" || env == "dev"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	env := getEnv("APP_ENV", "development")
	return env == "production" || env == "prod"
}
