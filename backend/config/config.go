package config

import (
	"bufio"
	"os"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	CORS     CORSConfig
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port     string
	CertFile string
	KeyFile  string
}

// DatabaseConfig holds PostgreSQL connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// CORSConfig holds CORS settings
type CORSConfig struct {
	AllowedOrigins []string
}

// Load reads configuration from environment variables with defaults
func Load() *Config {
	// Try to load .env file
	loadEnvFile(".env")

	return &Config{
		Server: ServerConfig{
			Port:     getEnv("SERVER_PORT", "8080"),
			CertFile: getEnv("TLS_CERT_FILE", ""),
			KeyFile:  getEnv("TLS_KEY_FILE", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "kong"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "smashqueue"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			SecretKey:            getEnv("AUTH_SECRET_KEY", "smashqueue-secret-key-change-in-production-32bytes!"),
			AccessTokenDuration:  30 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{
				getEnv("CORS_ORIGIN", "http://localhost:3000"),
			},
		},
	}
}

// loadEnvFile reads a .env file and sets environment variables
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // File doesn't exist, use defaults
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first =
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConnectionString builds the PostgreSQL connection string
func (d *DatabaseConfig) ConnectionString() string {
	return "postgresql://" + d.User + ":" + d.Password + "@" + d.Host + ":" + d.Port + "/" + d.Name + "?sslmode=" + d.SSLMode
}
