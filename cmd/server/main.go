package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup structured logging
	logger := SetupLogger(config)

	// Log startup information
	logger.Info("Starting Memos server",
		"env", config.Env,
		"port", config.Port,
		"allowed_hosts", config.AllowedHosts,
	)

	// Set Gin mode based on environment
	if config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Configure trusted proxies for Fly.io and local development
	if err := router.SetTrustedProxies(config.TrustedProxies); err != nil {
		logger.Error("Failed to set trusted proxies", "error", err)
		log.Fatalf("Fatal: %v", err)
	}
	logger.Debug("Configured trusted proxies", "proxies", config.TrustedProxies)

	// Configure security middleware
	secureMiddleware := secure.New(secure.Config{
		AllowedHosts:          config.AllowedHosts,
		SSLRedirect:           config.Env == "production",
		STSSeconds:            int64(config.HSTSMaxAge),
		STSIncludeSubdomains:  true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: buildCSP(config.CSPMode),
	})
	router.Use(secureMiddleware)
	logger.Debug("Configured security middleware",
		"hsts_enabled", config.Env == "production",
		"csp_mode", config.CSPMode,
	)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "memos",
		})
	})

	// Reserved API namespace for future development
	// api := router.Group("/api/v1")
	// {
	// 	api.GET("/health", func(c *gin.Context) {
	// 		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	// 	})
	// }

	// Serve static files from Hugo's public directory
	// TODO: Replace with router.Static() after security middleware
	router.NoRoute(func(c *gin.Context) {
		path := "./public" + c.Request.URL.Path
		c.File(path)
	})

	// Start server
	logger.Info("Server listening", "port", config.Port)
	if err := router.Run(":" + config.Port); err != nil {
		logger.Error("Failed to start server", "error", err)
		log.Fatalf("Fatal: %v", err)
	}
}
