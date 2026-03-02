// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl          string
	JWTSecret      string
	PlaidClientID  string
	PlaidSecret    string
	PlaidEnv       string
	Port           string
	ResetDB        bool
	AllowedOrigins []string
	BaseURL        string
	WebhookURL     string
}

func Load() (*Config, error) {
	// .env is optional — env vars may come directly (e.g. Docker Compose)
	_ = godotenv.Load()

	if os.Getenv("DATABASE_URL") == "" {
		return nil, fmt.Errorf("DATABASE_URL is required (set via .env file or environment variable)")
	}
	if os.Getenv("PLAID_CLIENT_ID") == "" {
		return nil, fmt.Errorf("PLAID_CLIENT_ID is required (set via .env file or environment variable)")
	}
	if os.Getenv("PLAID_SECRET") == "" {
		return nil, fmt.Errorf("PLAID_SECRET is required (set via .env file or environment variable)")
	}
	if os.Getenv("PLAID_ENV") == "" {
		return nil, fmt.Errorf("PLAID_ENV is required (set via .env file or environment variable)")
	}
	if os.Getenv("ALLOWED_ORIGINS") == "" {
		return nil, fmt.Errorf("ALLOWED_ORIGINS is required (set via .env file or environment variable)")
	}
	if os.Getenv("WEBHOOK_URL") == "" {
		return nil, fmt.Errorf("WEBHOOK_URL is required (set via .env file or environment variable)")
	}

	return &Config{
		DBUrl:          os.Getenv("DATABASE_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		PlaidClientID:  os.Getenv("PLAID_CLIENT_ID"),
		PlaidSecret:    os.Getenv("PLAID_SECRET"),
		PlaidEnv:       os.Getenv("PLAID_ENV"),
		Port:           os.Getenv("BACKEND_PORT"),
		ResetDB:        os.Getenv("RESET_DB") == "true",
		AllowedOrigins: strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
		BaseURL:        os.Getenv("BASE_URL") ,
		WebhookURL:     os.Getenv("WEBHOOK_URL"),
	}, nil

}
