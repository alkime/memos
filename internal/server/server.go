package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/alkime/memos/internal/config"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server.
type Server struct {
	config *config.Config
	logger *slog.Logger
	router *gin.Engine
}

// New creates a new Server instance.
func New(cfg *config.Config, logger *slog.Logger) *Server {
	// Set Gin mode based on environment
	if cfg.Env == config.EnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.Default()

	// Configure proxy trust for production (Fly.io)
	if cfg.Env == config.EnvProduction {
		router.TrustedPlatform = gin.PlatformFlyIO
		logger.Debug("Configured trusted platform", "platform", "fly.io")
	}
	// Development: no reverse proxy, uses direct client IP

	server := &Server{
		config: cfg,
		logger: logger,
		router: router,
	}

	// Setup middleware and routes
	setupSecurityMiddleware(router, cfg, logger)
	server.setupRoutes()

	return server
}

// Router returns the server's router for testing.
func (s *Server) Router() *gin.Engine {
	return s.router
}

// Run starts the HTTP server.
func Run(s *Server) error {
	s.logger.Info("Server listening", "port", s.config.Port)
	if err := s.router.Run(":" + s.config.Port); err != nil {
		return fmt.Errorf("failed to start server on port %s: %w", s.config.Port, err)
	}

	return nil
}

// setupRoutes configures all HTTP routes.
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.handleHealth)

	// Reserved API namespace for future development
	// api := s.router.Group("/api/v1")
	// {
	// 	api.GET("/health", s.handleHealth)
	// }

	// Serve static files from Hugo's public directory
	// Using gin-contrib/static for better integration with Gin middleware
	s.router.Use(static.Serve("/", static.LocalFile("./public", false)))
}

// handleHealth handles the health check endpoint.
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "memos",
	})
}
