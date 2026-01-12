package envx

import (
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env file
// In local development: uses .env file
// In cloud (Cloud Run): uses environment variables from Secret Manager (no .env file needed)
func LoadEnv() error {
	// Check if .env file exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		// .env is optional in production/cloud environments
		return nil
	}

	// Load .env file
	if err := godotenv.Load(); err != nil {
		return err
	}

	return nil
}

// GetEnv returns the current environment from ENV variable
func GetEnv() string {
	env := os.Getenv("ENV")
	if env == "" {
		return "development"
	}
	return env
}

// IsProduction returns true if running in production
func IsProduction() bool {
	return GetEnv() == "production"
}

// IsDevelopment returns true if running in development
func IsDevelopment() bool {
	return GetEnv() == "development"
}
