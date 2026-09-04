package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	// Clear relevant env vars so we test defaults cleanly.
	vars := []string{
		"HTTP_PORT",
		"READ_TIMEOUT_SECONDS",
		"WRITE_TIMEOUT_SECONDS",
		"SHUTDOWN_TIMEOUT_SECONDS",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}

	cfg := Load()

	if cfg.HTTPPort != "8080" {
		t.Errorf("HTTPPort: got %q, want %q", cfg.HTTPPort, "8080")
	}
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("ReadTimeout: got %v, want %v", cfg.ReadTimeout, 10*time.Second)
	}
	if cfg.WriteTimeout != 10*time.Second {
		t.Errorf("WriteTimeout: got %v, want %v", cfg.WriteTimeout, 10*time.Second)
	}
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Errorf("ShutdownTimeout: got %v, want %v", cfg.ShutdownTimeout, 30*time.Second)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("HTTP_PORT", "9090")
	os.Setenv("READ_TIMEOUT_SECONDS", "5")
	os.Setenv("WRITE_TIMEOUT_SECONDS", "15")
	os.Setenv("SHUTDOWN_TIMEOUT_SECONDS", "60")
	defer func() {
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("READ_TIMEOUT_SECONDS")
		os.Unsetenv("WRITE_TIMEOUT_SECONDS")
		os.Unsetenv("SHUTDOWN_TIMEOUT_SECONDS")
	}()

	cfg := Load()

	if cfg.HTTPPort != "9090" {
		t.Errorf("HTTPPort: got %q, want %q", cfg.HTTPPort, "9090")
	}
	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout: got %v, want %v", cfg.ReadTimeout, 5*time.Second)
	}
	if cfg.WriteTimeout != 15*time.Second {
		t.Errorf("WriteTimeout: got %v, want %v", cfg.WriteTimeout, 15*time.Second)
	}
	if cfg.ShutdownTimeout != 60*time.Second {
		t.Errorf("ShutdownTimeout: got %v, want %v", cfg.ShutdownTimeout, 60*time.Second)
	}
}

func TestGetDurationInvalidFallsBack(t *testing.T) {
	os.Setenv("READ_TIMEOUT_SECONDS", "not-a-number")
	defer os.Unsetenv("READ_TIMEOUT_SECONDS")

	cfg := Load()
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("invalid env should fall back to default; got %v", cfg.ReadTimeout)
	}
}

func TestGetDurationZeroFallsBack(t *testing.T) {
	os.Setenv("READ_TIMEOUT_SECONDS", "0")
	defer os.Unsetenv("READ_TIMEOUT_SECONDS")

	cfg := Load()
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("zero duration should fall back to default; got %v", cfg.ReadTimeout)
	}
}
