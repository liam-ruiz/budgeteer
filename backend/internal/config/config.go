// internal/config/config.go
package config

import (
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
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
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
	}, nil
}
