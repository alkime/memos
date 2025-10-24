# Security Hardening Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement OWASP baseline security hardening for the Memos Go/Gin web server with environment-aware configuration, security headers, and structured logging.

**Architecture:** Add configuration layer using godotenv/envconfig, implement gin-contrib/secure middleware for security headers, configure trusted proxies for Fly.io, replace manual file serving with router.Static(), and add structured logging with slog.

**Tech Stack:** Go 1.23, Gin, gin-contrib/secure, godotenv, envconfig, slog (stdlib)

---

## Progress Tracker

**Status:** In Progress (8/17 tasks completed - 47%)
**Last Updated:** 2025-10-23

### Core Implementation (Tasks 1-8) ✅
- [x] Task 1: Add Dependencies
- [x] Task 2: Create Configuration Structure
- [x] Task 3: Create Example Environment File
- [x] Task 4: Add Structured Logging Setup
- [x] Task 5: Update Main to Use Configuration
- [x] Task 6: Configure Trusted Proxies
- [x] Task 7: Add Security Middleware
- [x] Task 8: Replace Manual File Serving with Secure NoRoute

### Testing & Documentation (Tasks 9-13) 🔄
- [ ] Task 9: Create Development Environment File (.env)
- [ ] Task 10: Manual Security Header Verification
- [ ] Task 11: Test Health Endpoint
- [ ] Task 12: Build and Test Binary
- [ ] Task 13: Update CLAUDE.md Documentation

### Verification & Deployment (Tasks 14-17) ⏳
- [ ] Task 14: Verify Build Process
- [ ] Task 15: Final Integration Test
- [ ] Task 16: Create Deployment Checklist
- [ ] Task 17: Final Commit and Summary

**Key Achievements:**
- ✅ No Gin security warnings
- ✅ All OWASP baseline security headers implemented
- ✅ Path traversal protection via http.FileServer
- ✅ Trusted proxy configuration for Fly.io
- ✅ Structured JSON logging with slog
- ✅ Environment-aware configuration (dev/prod)

---

## Task 1: Add Dependencies

**Files:**
- Modify: `go.mod` (managed by go get)
- Modify: `go.sum` (managed by go get)

**Step 1: Add godotenv dependency**

Run: `go get github.com/joho/godotenv`
Expected: Dependency added to go.mod

**Step 2: Add envconfig dependency**

Run: `go get github.com/kelseyhightower/envconfig`
Expected: Dependency added to go.mod

**Step 3: Add gin-contrib/secure dependency**

Run: `go get github.com/gin-contrib/secure`
Expected: Dependency added to go.mod

**Step 4: Verify dependencies**

Run: `go mod tidy`
Expected: Clean output, all dependencies resolved

**Step 5: Commit dependencies**

```bash
git add go.mod go.sum
git commit -m "build: add security and config dependencies"
```

---

## Task 2: Create Configuration Structure

**Files:**
- Create: `cmd/server/config.go`
- Test: Manual verification via compilation

**Step 1: Write configuration struct**

Create `cmd/server/config.go`:

```go
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
```

**Step 2: Verify config compiles**

Run: `go build ./cmd/server`
Expected: Clean compilation, binary created

**Step 3: Commit configuration structure**

```bash
git add cmd/server/config.go
git commit -m "feat: add configuration structure with env loading"
```

---

## Task 3: Create Example Environment File

**Files:**
- Create: `.env.example`

**Step 1: Create .env.example**

Create `.env.example`:

```bash
# Server Configuration
ENV=development
PORT=8080

# Security Configuration
ALLOWED_HOSTS=localhost,alkime-memos.fly.dev
TRUSTED_PROXIES=10.0.0.0/8,172.16.0.0/12
HSTS_MAX_AGE=31536000
CSP_MODE=relaxed

# Logging Configuration
LOG_LEVEL=info
```

**Step 2: Verify .env is in .gitignore**

Run: `grep "^\.env$" .gitignore`
Expected: `.env` appears in output (already in .gitignore)

**Step 3: Commit example file**

```bash
git add .env.example
git commit -m "docs: add example environment configuration"
```

---

## Task 4: Add Structured Logging Setup

**Files:**
- Create: `cmd/server/logging.go`

**Step 1: Create logging setup**

Create `cmd/server/logging.go`:

```go
package main

import (
	"log/slog"
	"os"
)

// SetupLogger configures structured logging based on environment
func SetupLogger(config *Config) *slog.Logger {
	// Determine log level
	logLevel := slog.LevelInfo
	if config.Env == "development" {
		logLevel = slog.LevelDebug
	}
	if config.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}

	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	logger := slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)

	return logger
}
```

**Step 2: Verify logging compiles**

Run: `go build ./cmd/server`
Expected: Clean compilation

**Step 3: Commit logging setup**

```bash
git add cmd/server/logging.go
git commit -m "feat: add structured logging with slog"
```

---

## Task 5: Update Main to Use Configuration

**Files:**
- Modify: `cmd/server/main.go:1-56`

**Step 1: Replace main.go with config-aware version**

Replace contents of `cmd/server/main.go`:

```go
package main

import (
	"log"
	"log/slog"
	"net/http"

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
```

**Step 2: Test server starts with defaults**

Run: `go run cmd/server/main.go`
Expected: Server starts on port 8080 with JSON-formatted logs

**Step 3: Stop server and commit**

Press Ctrl+C to stop server.

```bash
git add cmd/server/main.go
git commit -m "refactor: integrate configuration and logging into main"
```

---

## Task 6: Configure Trusted Proxies

**Files:**
- Modify: `cmd/server/main.go:35-36` (after router creation)

**Step 1: Add trusted proxy configuration**

In `cmd/server/main.go`, add after line 36 (`router := gin.Default()`):

```go
	// Create Gin router
	router := gin.Default()

	// Configure trusted proxies for Fly.io and local development
	if err := router.SetTrustedProxies(config.TrustedProxies); err != nil {
		logger.Error("Failed to set trusted proxies", "error", err)
		log.Fatalf("Fatal: %v", err)
	}
	logger.Debug("Configured trusted proxies", "proxies", config.TrustedProxies)
```

**Step 2: Test trusted proxy configuration**

Run: `go run cmd/server/main.go`
Expected: No warnings about "trusted all proxies", debug log shows configured proxies

**Step 3: Stop server and commit**

Press Ctrl+C to stop server.

```bash
git add cmd/server/main.go
git commit -m "feat: configure trusted proxies for Fly.io"
```

---

## Task 7: Add Security Middleware

**Files:**
- Modify: `cmd/server/main.go:43-44` (after trusted proxies)

**Step 1: Add security middleware configuration**

In `cmd/server/main.go`, add after the trusted proxies configuration:

```go
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
```

And add the import at the top:

```go
import (
	"log"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)
```

**Step 2: Test security middleware**

Run: `go run cmd/server/main.go`
Expected: Server starts with security middleware configured

**Step 3: Stop server and commit**

Press Ctrl+C to stop server.

```bash
git add cmd/server/main.go
git commit -m "feat: add security headers middleware"
```

---

## Task 8: Replace Manual File Serving with router.Static()

**Files:**
- Modify: `cmd/server/main.go:71-75` (NoRoute section)

**Step 1: Replace NoRoute with router.Static()**

Replace the NoRoute section with:

```go
	// Serve static files from Hugo's public directory
	// Using router.Static() for built-in path traversal protection
	router.Static("/", "./public")
```

**Step 2: Test static file serving**

Run: `go run cmd/server/main.go`
Expected: Server starts successfully

**Step 3: Test a request (if public/ has files)**

In another terminal:
Run: `curl -I http://localhost:8080/`
Expected: Response with security headers

**Step 4: Stop server and commit**

Press Ctrl+C to stop server.

```bash
git add cmd/server/main.go
git commit -m "fix: replace manual file serving with router.Static for security"
```

---

## Task 9: Create Development Environment File

**Files:**
- Create: `.env` (gitignored)

**Step 1: Create local .env for testing**

Create `.env`:

```bash
ENV=development
PORT=8080
ALLOWED_HOSTS=localhost
TRUSTED_PROXIES=172.16.0.0/12,127.0.0.1
CSP_MODE=relaxed
LOG_LEVEL=debug
```

**Step 2: Test .env loading**

Run: `go run cmd/server/main.go`
Expected: Server starts with debug logging, shows loaded configuration

**Step 3: Verify configuration in logs**

Check logs for:
- `"env":"development"`
- `"allowed_hosts":["localhost"]`
- `"level":"DEBUG"` messages appear

**Step 4: Stop server**

Press Ctrl+C (no commit - .env is gitignored)

---

## Task 10: Manual Security Header Verification

**Files:**
- None (testing only)

**Step 1: Start server in development mode**

Run: `ENV=development go run cmd/server/main.go`

**Step 2: Test security headers**

In another terminal:
Run: `curl -I http://localhost:8080/`

Expected headers present:
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy: default-src 'self'...`
- `Referrer-Policy: strict-origin-when-cross-origin`

Expected headers ABSENT (dev mode):
- `Strict-Transport-Security` (only in production)

**Step 3: Test production mode headers**

Stop server, then run:
Run: `ENV=production PORT=8080 ALLOWED_HOSTS=localhost go run cmd/server/main.go`

In another terminal:
Run: `curl -I http://localhost:8080/`

Expected additional header:
- `Strict-Transport-Security: max-age=31536000; includeSubdomains`

**Step 4: Test host validation**

Run: `curl -I -H "Host: evil.com" http://localhost:8080/`
Expected: Connection rejected or 403 response

**Step 5: Test path traversal protection**

Run: `curl -I http://localhost:8080/../../../etc/passwd`
Expected: 404 Not Found (not actual file contents)

**Step 6: Stop server**

Press Ctrl+C to stop

---

## Task 11: Test Health Endpoint

**Files:**
- None (testing only)

**Step 1: Start server**

Run: `go run cmd/server/main.go`

**Step 2: Test health endpoint**

Run: `curl http://localhost:8080/health`
Expected output:
```json
{"service":"memos","status":"healthy"}
```

**Step 3: Verify health endpoint has security headers**

Run: `curl -I http://localhost:8080/health`
Expected: All security headers present on health endpoint

**Step 4: Stop server**

Press Ctrl+C

---

## Task 12: Build and Test Binary

**Files:**
- Generate: `build/server` (gitignored)

**Step 1: Build binary**

Run: `go build -o build/server ./cmd/server`
Expected: Binary created at `build/server`

**Step 2: Test binary runs**

Run: `./build/server`
Expected: Server starts, loads config from .env

**Step 3: Test binary respects environment**

Stop server.
Run: `ENV=production PORT=9000 ./build/server`
Expected: Server starts on port 9000, production mode

**Step 4: Stop server and clean build**

Press Ctrl+C, then:
Run: `rm -rf build/`

---

## Task 13: Update CLAUDE.md Documentation

**Files:**
- Modify: `CLAUDE.md` (append to Environment Variables section)

**Step 1: Add environment variables documentation**

Update the "Environment Variables" section in `CLAUDE.md`:

```markdown
## Environment Variables

### Server Configuration
- `ENV` - Environment name: `development` or `production` (default: `development`)
- `PORT` - Server port (default: `8080`)
- `LOG_LEVEL` - Logging verbosity: `debug` or `info` (default: `info`)

### Security Configuration
- `ALLOWED_HOSTS` - Comma-separated list of allowed Host header values (default: `localhost,alkime-memos.fly.dev`)
- `TRUSTED_PROXIES` - Comma-separated list of trusted proxy CIDR ranges (default: `10.0.0.0/8,172.16.0.0/12`)
- `HSTS_MAX_AGE` - HSTS max-age in seconds, production only (default: `31536000`)
- `CSP_MODE` - Content Security Policy mode: `relaxed` or `strict` (default: `relaxed`)

### Local Development

Create a `.env` file in the project root (gitignored) with development settings. See `.env.example` for all available options.

### Production Deployment (Fly.io)

Set environment variables using Fly.io secrets:

```bash
fly secrets set ENV=production
fly secrets set ALLOWED_HOSTS=alkime-memos.fly.dev,memos.alkime.dev
fly secrets set CSP_MODE=strict
```

The `PORT` variable is automatically set by Fly.io.
```

**Step 2: Commit documentation**

```bash
git add CLAUDE.md
git commit -m "docs: document security configuration environment variables"
```

---

## Task 14: Verify Build Process

**Files:**
- None (verification only)

**Step 1: Clean any generated files**

Run: `go clean`
Expected: Clean output

**Step 2: Verify go.mod is clean**

Run: `go mod tidy`
Expected: No changes

**Step 3: Run full build**

Run: `go build ./...`
Expected: All packages compile successfully

**Step 4: Verify no git changes**

Run: `git status`
Expected: Clean working tree

---

## Task 15: Final Integration Test

**Files:**
- None (integration test)

**Step 1: Start server with production-like config**

Create temporary test .env:
```bash
cat > .env.test << 'EOF'
ENV=production
PORT=8080
ALLOWED_HOSTS=localhost
TRUSTED_PROXIES=127.0.0.1
CSP_MODE=strict
HSTS_MAX_AGE=31536000
LOG_LEVEL=info
EOF
```

Run: `ENV=production LOG_LEVEL=info go run cmd/server/main.go`

**Step 2: Verify startup logs**

Check logs show:
- JSON format
- `"env":"production"`
- `"level":"INFO"`
- No warnings about trusted proxies
- Security middleware configured

**Step 3: Test all security headers**

Run: `curl -v http://localhost:8080/ 2>&1 | grep -E "^< (X-|Strict-Transport|Content-Security|Referrer)"`

Expected headers:
- X-Frame-Options: DENY
- X-Content-Type-Options: nosniff
- X-XSS-Protection: 1; mode=block
- Strict-Transport-Security: max-age=31536000; includeSubdomains
- Content-Security-Policy: (strict mode policy)
- Referrer-Policy: strict-origin-when-cross-origin

**Step 4: Test health endpoint**

Run: `curl http://localhost:8080/health | jq`
Expected: Valid JSON with status=healthy

**Step 5: Stop server and clean up**

Press Ctrl+C, then:
Run: `rm .env.test`

---

## Task 16: Create Deployment Checklist

**Files:**
- Create: `docs/deployment/fly-io-security-config.md`

**Step 1: Create deployment documentation**

Create `docs/deployment/fly-io-security-config.md`:

```markdown
# Fly.io Security Configuration

## Required Environment Variables

Set these secrets before deploying:

```bash
# Required
fly secrets set ENV=production
fly secrets set ALLOWED_HOSTS=alkime-memos.fly.dev,memos.alkime.dev
fly secrets set CSP_MODE=strict

# Optional (defaults are production-ready)
fly secrets set HSTS_MAX_AGE=31536000
fly secrets set TRUSTED_PROXIES=10.0.0.0/8
fly secrets set LOG_LEVEL=info
```

## Verification After Deployment

1. **Check server logs:**
   ```bash
   fly logs
   ```
   Verify:
   - No warnings about trusted proxies
   - JSON-formatted logs
   - "env":"production"

2. **Test security headers:**
   ```bash
   curl -I https://alkime-memos.fly.dev/
   ```
   Verify presence of:
   - Strict-Transport-Security
   - X-Frame-Options
   - X-Content-Type-Options
   - Content-Security-Policy

3. **Test health endpoint:**
   ```bash
   curl https://alkime-memos.fly.dev/health
   ```
   Expected: `{"status":"healthy","service":"memos"}`

## Security Header Testing Tools

- [securityheaders.com](https://securityheaders.com/)
- [Mozilla Observatory](https://observatory.mozilla.org/)

Test your deployment:
```
https://securityheaders.com/?q=alkime-memos.fly.dev
```
```

**Step 2: Commit deployment documentation**

```bash
mkdir -p docs/deployment
git add docs/deployment/fly-io-security-config.md
git commit -m "docs: add Fly.io security configuration guide"
```

---

## Task 17: Final Commit and Summary

**Files:**
- None (review only)

**Step 1: Review all changes**

Run: `git log --oneline --graph -20`
Expected: See all commits from this implementation

**Step 2: Verify no uncommitted changes**

Run: `git status`
Expected: Clean working tree

**Step 3: Review files changed**

Run: `git diff main...HEAD --stat`
Expected: Shows all modified files

**Step 4: Create summary**

Document what was implemented:
- ✅ Configuration management (godotenv + envconfig)
- ✅ Trusted proxy configuration for Fly.io
- ✅ Security headers middleware (gin-contrib/secure)
- ✅ Structured logging (slog)
- ✅ Secure static file serving (router.Static)
- ✅ Environment-aware settings (dev/prod)
- ✅ Documentation updates
- ✅ Deployment guide

---

## Verification Commands Summary

**Development testing:**
```bash
# Start server
go run cmd/server/main.go

# Test headers
curl -I http://localhost:8080/

# Test health
curl http://localhost:8080/health
```

**Production simulation:**
```bash
ENV=production PORT=8080 ALLOWED_HOSTS=localhost CSP_MODE=strict go run cmd/server/main.go
```

**Build verification:**
```bash
go build -o build/server ./cmd/server
./build/server
```

## Next Steps After Implementation

1. Test deployment to Fly.io staging environment
2. Configure Fly.io secrets per deployment guide
3. Run security header scan on deployed site
4. Monitor logs for any configuration issues
5. Consider adding rate limiting middleware (future task)
6. Plan API authentication implementation (future phase)

## Related Skills

- @superpowers:verification-before-completion - Use before claiming tasks complete
- @superpowers:systematic-debugging - If issues arise during testing
- @superpowers:requesting-code-review - After implementation before merge
