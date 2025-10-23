package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration
type Config struct {
	// Server settings
	Env  string `envconfig:"ENV" default:"development"`
	Port string `envconfig:"PORT" default:"8080"`

	// Security settings
	AllowedHosts   []string `envconfig:"ALLOWED_HOSTS" default:"localhost,alkime-memos.fly.dev"`
	TrustedProxies []string `envconfig:"TRUSTED_PROXIES" default:"10.0.0.0/8,172.16.0.0/12"`
	HSTSMaxAge     int      `envconfig:"HSTS_MAX_AGE" default:"31536000"`
	CSPMode        string   `envconfig:"CSP_MODE" default:"relaxed"`

	// Logging settings
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

// LoadConfig loads configuration from .env file and environment variables
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
		return nil, err
	}

	return &config, nil
}

// buildCSP constructs Content Security Policy based on mode
func buildCSP(mode string) string {
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
