package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	// EnvProduction represents the production environment.
	EnvProduction = "production"
)

// Config holds all application configuration.
type Config struct {
	// Server settings
	Env  string `envconfig:"ENV" default:"development"`
	Port string `envconfig:"PORT" default:"8080"`

	// Security settings
	HSTSMaxAge int    `envconfig:"HSTS_MAX_AGE" default:"31536000"`
	CSPMode    string `envconfig:"CSP_MODE" default:"relaxed"`

	// Logging settings
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

// LoadConfig loads configuration from .env file and environment variables.
func LoadConfig() (*Config, error) {
	// Try to load .env file (optional for development)
	if err := godotenv.Load(); err != nil {
		// Not an error if file doesn't exist (expected in production)
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	// Parse environment variables into config struct
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}

	return &config, nil
}

// BuildCSP constructs Content Security Policy based on mode.
func BuildCSP(mode string) string {
	if mode == "strict" {
		// Production CSP
		return "default-src 'self'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"script-src 'self'; " +
			"img-src 'self' https://*.tigris.dev data:; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
	}

	// Development/relaxed CSP
	return "default-src 'self'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"script-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data:"
}
