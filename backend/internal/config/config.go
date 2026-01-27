package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds the application configuration
type Config struct {
	// AppEnv is the application environment (development, production, etc.)
	AppEnv string `envconfig:"APP_ENV" default:"development"`

	// Port is the HTTP server port
	Port int `envconfig:"API_PORT" default:"8080"`

	// DatabaseURL is the PostgreSQL connection string
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// AllowedOrigins is a comma-separated list of allowed CORS origins
	// In production, this should be set to specific domains
	AllowedOrigins string `envconfig:"ALLOWED_ORIGINS" default:""`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
