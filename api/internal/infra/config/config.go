package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port  string
	DBDSN string
}

func Load() *Config {
	// Loads api/.env when present (e.g. local `go run`); real environment
	// variables (docker-compose, CI, ...) always take precedence.
	_ = godotenv.Load()

	return &Config{
		Port:  getEnv("PORT", "8080"),
		DBDSN: buildDBDSN(),
	}
}

// buildDBDSN mirrors the POSTGRES_* variables docker-compose.yml feeds into
// the database container, so both point at the same database.
func buildDBDSN() string {
	host := getEnv("DATABASE_HOST", "localhost")
	port := getEnv("DATABASE_PORT", "5432")
	user := getEnv("DATABASE_USERNAME", "postgres")
	password := getEnv("DATABASE_PASSWORD", "postgres")
	name := getEnv("DATABASE_NAME", "food_store") + os.Getenv("APP_ENV")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, name)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
