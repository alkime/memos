package server

import (
	"log/slog"
	"net/http"

	"github.com/alkime/memos/internal/config"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	logger *slog.Logger
	router *gin.Engine
}

// New creates a new Server instance
func New(cfg *config.Config, logger *slog.Logger) *Server {
	// Set Gin mode based on environment
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.Default()

	// Configure proxy trust for production (Fly.io)
	if cfg.Env == "production" {
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

// Run starts the HTTP server
func Run(s *Server) error {
	s.logger.Info("Server listening", "port", s.config.Port)
	return s.router.Run(":" + s.config.Port)
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.handleHealth)

	// Reserved API namespace for future development
	// api := s.router.Group("/api/v1")
	// {
	// 	api.GET("/health", s.handleHealth)
	// }

	// Serve static files from Hugo's public directory as fallback
	// Using http.FileServer for built-in path traversal protection
	// NoRoute only triggers when no explicit routes match (like /health)
	s.router.NoRoute(gin.WrapH(http.FileServer(http.Dir("./public"))))
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "memos",
	})
}
