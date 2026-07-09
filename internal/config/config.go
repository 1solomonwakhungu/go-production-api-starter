package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration values sourced from the environment.
type Config struct {
	ServerPort      string
	Environment     string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	JWTSecret       string
	JWTExpiryHours  int
	LogLevel        string
	AllowedOrigins  []string
}

// Load reads configuration from environment variables, applying sensible defaults.
func Load() (*Config, error) {
	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS: %w", err)
	}

	cfg := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		Environment:    getEnv("APP_ENV", "development"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "appdb"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiryHours: jwtExpiry,
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"*"}),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate ensures required fields are set for production.
func (c *Config) Validate() error {
	if c.Environment == "production" {
		if c.JWTSecret == "change-me-in-production" {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
		if c.DBPassword == "postgres" {
			return fmt.Errorf("DB_PASSWORD must be set in production")
		}
	}
	return nil
}

// DSN returns a Postgres connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// IsProduction returns true if the app is running in production mode.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if the app is running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var result []string
	for _, s := range splitComma(v) {
		if s != "" {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func splitComma(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
