package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// PostgresConfig holds database connection parameters
type PostgresConfig struct {
	host     string
	user     string
	password string
	name     string
	ssl      string
	port     int
}

// String returns a formatted connection string for database initialization
func (p PostgresConfig) String() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		p.host,
		p.user,
		p.password,
		p.name,
		p.port,
		p.ssl,
	)
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port string
}

// GameConfig defines the betting constraints
type GameConfig struct {
	MinBetAmount float64
	MaxBetAmount float64
}

// Config aggregates all application configuration categories
type Config struct {
	Postgres PostgresConfig
	Server   ServerConfig
	Game     GameConfig
}

// New initializes configuration with environment variables or defaults
func New() *Config {
	return &Config{
		Postgres: PostgresConfig{
			host:     getEnv("DB_HOST", ""),
			user:     getEnv("DB_USER", ""),
			password: getEnv("DB_PASSWORD", ""),
			name:     getEnv("DB_NAME", ""),
			ssl:      getEnv("DB_SSL", "disable"),
			port:     getEnvAsInt("DB_PORT", 5432),
		},
		Server: ServerConfig{Port: getEnv("SERVER_PORT", "80")},
		Game: GameConfig{
			MinBetAmount: getEnvAsFloat("MIN_BET", 10.0),
			MaxBetAmount: getEnvAsFloat("MAX_BET", 100.0),
		},
	}
}

// getEnv retrieves environment variables with fallback to default values
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvAsInt parses integer environment variables with fallback
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// getEnvAsFloat parses float environment variables with logging on parse failures
func getEnvAsFloat(name string, defaultVal float64) float64 {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	log.Printf("could not parse %s to float, using default value of %f", name, defaultVal)
	return defaultVal
}
