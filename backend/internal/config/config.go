package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port           int      `envconfig:"PORT" default:"8080"`
	DatabaseURL    string   `envconfig:"DATABASE_URL" required:"true"`
	LogLevel       string   `envconfig:"LOG_LEVEL" default:"info"`
	AllowedOrigins []string `envconfig:"ALLOWED_ORIGINS" default:"*"`
	AutoMigrate    bool     `envconfig:"AUTO_MIGRATE" default:"false"`
}

// Load reads configuration from environment variables.
func Load() Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}
