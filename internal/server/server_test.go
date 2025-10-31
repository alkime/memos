package server_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alkime/memos/internal/config"
	"github.com/alkime/memos/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	// Setup: Create server with test config
	cfg := &config.Config{
		Env:        "test",
		Port:       "8080",
		HSTSMaxAge: 31536000,
		CSPMode:    "relaxed",
		LogLevel:   "info",
	}

	// Create a test logger (discard output)
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level:       slog.LevelError, // Only show errors during tests
		AddSource:   false,
		ReplaceAttr: nil,
	}))

	srv := server.New(cfg, logger)

	// Create test HTTP request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute request
	srv.Router().ServeHTTP(w, req)

	// Assert response using testify
	assert.Equal(t, http.StatusOK, w.Code, "Health endpoint should return 200 OK")
	assert.Contains(t, w.Body.String(), "healthy", "Response should contain 'healthy'")
	assert.Contains(t, w.Body.String(), "memos", "Response should contain service name 'memos'")
}
