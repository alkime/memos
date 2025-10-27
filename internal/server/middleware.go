// Package server provides HTTP server and middleware functionality.
package server

import (
	"log/slog"

	"github.com/alkime/memos/internal/config"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

// setupSecurityMiddleware configures and applies security middleware to the router.
func setupSecurityMiddleware(router *gin.Engine, cfg *config.Config, logger *slog.Logger) {
	// Configure HSTS for production only
	stsSeconds := int64(0)
	if cfg.Env == config.EnvProduction {
		stsSeconds = int64(cfg.HSTSMaxAge)
	}

	// Create and apply security middleware
	//nolint:exhaustruct // Using default values for other secure.Config fields
	secureMiddleware := secure.New(secure.Config{
		STSSeconds:            stsSeconds,
		STSIncludeSubdomains:  true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: config.BuildCSP(cfg.CSPMode),
	})
	router.Use(secureMiddleware)

	logger.Debug("Configured security middleware",
		"hsts_enabled", cfg.Env == config.EnvProduction,
		"csp_mode", cfg.CSPMode,
	)
}
