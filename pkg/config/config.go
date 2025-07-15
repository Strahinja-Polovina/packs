package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	App      AppConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Name string
	Port int
	Mode string // gin mode: debug, release, test
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	LogLevel string
	Version  string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	// Load .env file if it exists (ignore errors if file doesn't exist)
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Name: getEnv("SERVER_NAME", "PacksAPI"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Mode: getEnv("GIN_MODE", "release"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "packs_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		App: AppConfig{
			LogLevel: getEnv("LOG_LEVEL", "INFO"),
			Version:  getEnv("APP_VERSION", "1.0.0"),
		},
	}
}

// ConnectionString returns the database connection string
func (c *DatabaseConfig) ConnectionString() string {
	return "host=" + c.Host + " port=" + c.Port + " user=" + c.User +
		" password=" + c.Password + " dbname=" + c.DBName + " sslmode=" + c.SSLMode
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
