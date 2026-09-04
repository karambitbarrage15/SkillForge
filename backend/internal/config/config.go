// Package config loads StreamForge configuration from environment variables.
// All fields have sensible defaults so the binary runs without any env vars set.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the runtime configuration for a StreamForge process.
// Later phases will extend this struct with database, Redis, and worker settings.
type Config struct {
	// HTTP
	HTTPPort        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// Worker
	HeartbeatInterval time.Duration
	HeartbeatTTL      time.Duration
}

// Load reads configuration from environment variables and returns a Config.
// Missing variables fall back to safe defaults.
func Load() Config {
	return Config{
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		ReadTimeout:       getDuration("READ_TIMEOUT_SECONDS", 10),
		ShutdownTimeout:   getDuration("SHUTDOWN_TIMEOUT_SECONDS", 30),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		RedisURL:          getEnv("REDIS_URL", ""),
		HeartbeatInterval: getDuration("HEARTBEAT_INTERVAL_SECONDS", 5),
		HeartbeatTTL:      getDuration("HEARTBEAT_TTL_SECONDS", 15),
	}
}

// getEnv returns the value of the environment variable key, or fallback if unset/empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getDuration reads an integer number of seconds from an env variable.
// Returns fallback * time.Second if the variable is unset or invalid.
func getDuration(key string, fallbackSeconds int) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(fallbackSeconds) * time.Second
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return time.Duration(fallbackSeconds) * time.Second
	}
	return time.Duration(n) * time.Second
}
