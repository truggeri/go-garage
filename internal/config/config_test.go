package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_WithDefaults(t *testing.T) {
	// Clear any environment variables that might interfere
	os.Clearenv()

	// Set required env
	t.Setenv("JWT_SECRET", "test-required-secret-key")

	config, err := Load()

	require.NoError(t, err, "Load should not return an error with valid defaults")
	assert.NotNil(t, config, "Config should not be nil")

	// Verify default values
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "./data/go-garage.db", config.Database.Path)
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "json", config.Logging.Format)
	assert.Equal(t, "test-required-secret-key", config.JWT.Secret)
	assert.Equal(t, "development", config.Env)
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	os.Clearenv()
	t.Setenv("SERVER_HOST", "localhost")
	t.Setenv("APP_PORT", "3000")
	t.Setenv("DB_PATH", "/var/data/app.db")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("JWT_SECRET", "test-secret-key")
	t.Setenv("ENVIRONMENT", "production")

	config, err := Load()

	require.NoError(t, err, "Load should not return an error with valid environment variables")
	assert.Equal(t, "localhost", config.Server.Host)
	assert.Equal(t, 3000, config.Server.Port)
	assert.Equal(t, "/var/data/app.db", config.Database.Path)
	assert.Equal(t, "debug", config.Logging.Level)
	assert.Equal(t, "text", config.Logging.Format)
	assert.Equal(t, "test-secret-key", config.JWT.Secret)
	assert.Equal(t, "production", config.Env)
}

func TestLoad_InvalidPort(t *testing.T) {
	os.Clearenv()
	t.Setenv("APP_PORT", "99999")

	config, err := Load()

	assert.Error(t, err, "Load should return an error with invalid port")
	assert.Nil(t, config, "Config should be nil when validation fails")
	assert.Contains(t, err.Error(), "invalid server port")
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	os.Clearenv()
	t.Setenv("LOG_LEVEL", "invalid")

	config, err := Load()

	assert.Error(t, err, "Load should return an error with invalid log level")
	assert.Nil(t, config, "Config should be nil when validation fails")
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestLoad_InvalidLogFormat(t *testing.T) {
	os.Clearenv()
	t.Setenv("LOG_FORMAT", "xml")

	config, err := Load()

	assert.Error(t, err, "Load should return an error with invalid log format")
	assert.Nil(t, config, "Config should be nil when validation fails")
	assert.Contains(t, err.Error(), "invalid log format")
}

func TestLoad_InvalidEnvironment(t *testing.T) {
	os.Clearenv()
	t.Setenv("ENVIRONMENT", "test")

	config, err := Load()

	assert.Error(t, err, "Load should return an error with invalid environment")
	assert.Nil(t, config, "Config should be nil when validation fails")
	assert.Contains(t, err.Error(), "invalid environment")
}

func TestLoad_ProductionWithoutJWTSecret(t *testing.T) {
	os.Clearenv()
	t.Setenv("ENVIRONMENT", "production")

	config, err := Load()

	assert.Error(t, err, "Load should return an error in production without JWT secret")
	assert.Nil(t, config, "Config should be nil when validation fails")
	assert.Contains(t, err.Error(), "JWT_SECRET is required")
}

func TestLoad_ProductionWithJWTSecret(t *testing.T) {
	os.Clearenv()
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("JWT_SECRET", "production-secret-key")

	config, err := Load()

	require.NoError(t, err, "Load should not return an error in production with JWT secret")
	assert.NotNil(t, config, "Config should not be nil")
	assert.Equal(t, "production", config.Env)
	assert.Equal(t, "production-secret-key", config.JWT.Secret)
}

func TestValidate_ValidPort(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Path: "./data/test.db",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		JWT: JWTConfig{
			Secret: "test-secret",
		},
		Env: "development",
	}

	err := config.Validate()

	assert.NoError(t, err, "Validation should pass with valid configuration")
}

func TestValidate_PortTooLow(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port: 0,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Env: "development",
	}

	err := config.Validate()

	assert.Error(t, err, "Validation should fail with port 0")
	assert.Contains(t, err.Error(), "invalid server port")
}

func TestValidate_PortTooHigh(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port: 65536,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Env: "development",
	}

	err := config.Validate()

	assert.Error(t, err, "Validation should fail with port 65536")
	assert.Contains(t, err.Error(), "invalid server port")
}

func TestIsDevelopment(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected bool
	}{
		{"Development environment", "development", true},
		{"Production environment", "production", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Env: tt.env}
			result := config.IsDevelopment()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected bool
	}{
		{"Development environment", "development", false},
		{"Production environment", "production", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Env: tt.env}
			result := config.IsProduction()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvOrDefault_WithValue(t *testing.T) {
	t.Setenv("TEST_KEY", "test_value")

	result := getEnvOrDefault("TEST_KEY", "default")

	assert.Equal(t, "test_value", result, "Should return environment variable value when set")
}

func TestGetEnvOrDefault_WithoutValue(t *testing.T) {
	result := getEnvOrDefault("NONEXISTENT_KEY", "default_value")

	assert.Equal(t, "default_value", result, "Should return default value when environment variable not set")
}

func TestGetEnvAsIntOrDefault_WithValidInt(t *testing.T) {
	t.Setenv("TEST_INT_KEY", "1234")

	result := getEnvAsIntOrDefault("TEST_INT_KEY", 999)

	assert.Equal(t, 1234, result, "Should return parsed integer value")
}

func TestGetEnvAsIntOrDefault_WithInvalidInt(t *testing.T) {
	t.Setenv("TEST_INT_KEY", "not_a_number")

	result := getEnvAsIntOrDefault("TEST_INT_KEY", 999)

	assert.Equal(t, 999, result, "Should return default value when parsing fails")
}

func TestGetEnvAsIntOrDefault_WithoutValue(t *testing.T) {
	result := getEnvAsIntOrDefault("NONEXISTENT_INT_KEY", 999)

	assert.Equal(t, 999, result, "Should return default value when environment variable not set")
}
