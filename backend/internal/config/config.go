package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration loaded from environment variables.
type Config struct {
	DatabaseURL     string
	JWTSecret       string
	BcryptCost      int
	AIRateLimitRPM  int
	AnthropicAPIKey string
	Port            string
}

// Load reads environment variables and returns a validated Config.
// Returns an error if any required variable is missing or invalid.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	bcryptCost := 12
	if v := os.Getenv("BCRYPT_COST"); v != "" {
		c, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("BCRYPT_COST must be an integer: %w", err)
		}
		if c < 10 {
			return nil, fmt.Errorf("BCRYPT_COST must be at least 10")
		}
		bcryptCost = c
	}

	aiRPM := 20
	if v := os.Getenv("AI_RATE_LIMIT_RPM"); v != "" {
		r, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("AI_RATE_LIMIT_RPM must be an integer: %w", err)
		}
		aiRPM = r
	}

	return &Config{
		DatabaseURL:     dbURL,
		JWTSecret:       jwtSecret,
		BcryptCost:      bcryptCost,
		AIRateLimitRPM:  aiRPM,
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		Port:            getEnvOrDefault("PORT", "8080"),
	}, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
