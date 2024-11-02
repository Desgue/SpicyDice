package config

import (
	"fmt"
	"os"
	"strconv"
)

type PostgresConfig struct {
	host     string
	user     string
	password string
	name     string
	ssl      string
	port     int
}

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

type Config struct {
	Postgres PostgresConfig
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		Postgres: PostgresConfig{
			host:     getEnv("DB_HOST", "db"),
			user:     getEnv("DB_USER", "postgres"),
			password: getEnv("DB_PASSWORD", ""),
			name:     getEnv("DB_NAME", "postgres"),
			ssl:      getEnv("DB_SSL", "disabled"),
			port:     getEnvAsInt("DB_PORT", 5432),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}
