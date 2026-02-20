package config

import "os"

type Config struct {
	HTTPAddr string
	DBPath   string
}

func Load() *Config {
	return &Config{
		HTTPAddr: envOr("CIELO_HTTP_ADDR", ":8080"),
		DBPath:   envOr("CIELO_DB_PATH", "cielo.db"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
