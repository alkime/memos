package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set Gin mode based on ENV variable
	env := os.Getenv("ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

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
	// NoRoute catches all unmatched routes and serves from public/
	router.NoRoute(func(c *gin.Context) {
		path := "./public" + c.Request.URL.Path
		c.File(path)
	})

	// Start server
	log.Printf("Starting Memos server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
