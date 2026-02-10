package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logging  LoggingConfig
	JWT      JWTConfig
	Env      string
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Path string
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret string
}

// Load creates a new Config by loading values from environment variables
// It automatically loads .env file if it exists (for local development)
func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Host: getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsIntOrDefault("APP_PORT", 8080),
		},
		Database: DatabaseConfig{
			Path: getEnvOrDefault("DB_PATH", "./data/go-garage.db"),
		},
		Logging: LoggingConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: getEnvOrDefault("LOG_FORMAT", "json"),
		},
		JWT: JWTConfig{
			Secret: getEnvOrDefault("JWT_SECRET", ""),
		},
		Env: getEnvOrDefault("ENVIRONMENT", "development"),
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate server port
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be between 1 and 65535)", c.Server.Port)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error)", c.Logging.Level)
	}

	// Validate log format
	validLogFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (must be one of: json, text)", c.Logging.Format)
	}

	// Validate environment
	validEnvironments := map[string]bool{
		"development": true,
		"production":  true,
	}
	if !validEnvironments[c.Env] {
		return fmt.Errorf("invalid environment: %s (must be one of: development, production)", c.Env)
	}

	// JWT secret is required in production
	if c.Env == "production" && c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required in production environment")
	}

	return nil
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsIntOrDefault gets an environment variable as an integer or returns a default value
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
