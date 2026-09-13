package config

import "os"

type Config struct {
	Port  string
	DBDSN string
}

func Load() *Config {
	return &Config{
		Port:  getEnv("PORT", "8080"),
		DBDSN: getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/food_store?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
