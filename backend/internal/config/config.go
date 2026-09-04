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
	WorkerCount         int
	WorkerConsumerGroup string
	HeartbeatInterval   time.Duration
	HeartbeatTTL        time.Duration
}

// Load reads configuration from environment variables and returns a Config.
// Missing variables fall back to safe defaults.
func Load() Config {
	return Config{
		HTTPPort:            getEnv("HTTP_PORT", "8080"),
		ReadTimeout:         getDuration("READ_TIMEOUT_SECONDS", 10),
		WriteTimeout:        getDuration("WRITE_TIMEOUT_SECONDS", 10),
		ShutdownTimeout:     getDuration("SHUTDOWN_TIMEOUT_SECONDS", 30),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		RedisURL:            getEnv("REDIS_URL", ""),
		WorkerCount:         getInt("WORKER_COUNT", 4),
		WorkerConsumerGroup: getEnv("WORKER_CONSUMER_GROUP", "streamforge_workers"),
		HeartbeatInterval:   getDuration("HEARTBEAT_INTERVAL_SECONDS", 5),
		HeartbeatTTL:        getDuration("HEARTBEAT_TTL_SECONDS", 15),
	}
}

// getEnv returns the value of the environment variable key, or fallback if unset/empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getInt reads an integer from an env variable.
func getInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

// getDuration reads an environment variable as a duration and falls back to a default.
func getDuration(key string, defaultSeconds int) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(defaultSeconds) * time.Second
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return time.Duration(defaultSeconds) * time.Second
	}
	return time.Duration(n) * time.Second
}
