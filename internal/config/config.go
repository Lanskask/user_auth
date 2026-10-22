// Package config loads service settings from environment variables (12-factor style).
package config

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Config holds every setting the server needs.
type Config struct {
	Addr         string
	DatabaseURL  string
	CookieSecure bool
	BcryptCost   int
	SessionTTL   time.Duration
}

// Load reads the configuration through getenv (usually os.Getenv) so tests can inject values.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Addr:         valueOr(getenv("AUTH_ADDR"), ":8080"),
		DatabaseURL:  getenv("AUTH_DATABASE_URL"),
		CookieSecure: true,
		BcryptCost:   12,
		SessionTTL:   24 * time.Hour,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("AUTH_DATABASE_URL is required")
	}

	if v := getenv("AUTH_COOKIE_SECURE"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return Config{}, fmt.Errorf("AUTH_COOKIE_SECURE: %w", err)
		}
		cfg.CookieSecure = b
	}

	if v := getenv("AUTH_BCRYPT_COST"); v != "" {
		c, err := strconv.Atoi(v)
		if err != nil || c < 10 || c > 14 {
			return Config{}, fmt.Errorf("AUTH_BCRYPT_COST must be an integer in [10, 14], got %q", v)
		}
		cfg.BcryptCost = c
	}

	if v := getenv("AUTH_SESSION_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("AUTH_SESSION_TTL must be a positive duration, got %q", v)
		}
		cfg.SessionTTL = d
	}

	return cfg, nil
}

func valueOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
