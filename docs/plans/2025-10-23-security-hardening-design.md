# Security Hardening Design for Memos Web Server

**Date:** 2025-10-23
**Status:** Approved
**Author:** Design session with Claude Code

## Executive Summary

This design implements OWASP baseline security hardening for the Memos Go/Gin web server. The implementation prepares the application for production deployment on Fly.io while laying the foundation for future API endpoints with authentication.

## Context & Requirements

### Current State
- Minimal Go/Gin web server serving Hugo static site
- Gin warning: "You trusted all proxies, this is NOT safe"
- No security headers configured
- Manual file serving with potential security gaps
- Basic logging with no structure

### Goals
- Fix proxy trust configuration warning
- Implement OWASP Top 10 baseline protections
- Prepare infrastructure for future API with authentication
- Enable multi-domain support (alkime-memos.fly.dev + memos.alki.me)
- Maintain compatibility with Fly.io Tigris object storage (future)

### Success Criteria
- No security warnings on server startup
- OWASP-recommended security headers present
- Environment-aware configuration (dev/prod)
- Structured logging for security events
- Clean foundation for adding authentication

## Architecture

### Component Overview

```
┌─────────────────────────────────────────────┐
│          Gin Engine (gin.Default())         │
├─────────────────────────────────────────────┤
│  1. SetTrustedProxies() Configuration       │
│     - Fly.io: 10.0.0.0/8                   │
│     - Local dev: 172.16.0.0/12             │
├─────────────────────────────────────────────┤
│  2. Security Middleware (gin-contrib/secure)│
│     - X-Frame-Options: DENY                 │
│     - X-Content-Type-Options: nosniff       │
│     - Strict-Transport-Security (prod)      │
│     - Content-Security-Policy               │
│     - Referrer-Policy                       │
│     - Allowed Hosts (env-configured)        │
├─────────────────────────────────────────────┤
│  3. Application Routes                      │
│     GET /health                             │
│     (future) /api/v1/*                      │
│     GET /* → Static("/", "./public")        │
└─────────────────────────────────────────────┘
```

### Data Flow

**Request Processing:**
1. Request arrives at Fly.io proxy
2. Gin validates proxy is trusted (SetTrustedProxies)
3. Security middleware adds headers and validates host
4. Route matching (health check → API → static files)
5. Response returned with security headers

## Detailed Design

### 1. Proxy Trust Configuration

**Implementation:**
```go
router := gin.Default()
router.SetTrustedProxies([]string{
    "10.0.0.0/8",      // Fly.io internal network
    "172.16.0.0/12",   // Local development
})
```

**Rationale:**
- Fly.io routes traffic through internal proxy layer using 10.x IPs
- Only trusted proxies can set X-Forwarded-For and related headers
- Prevents client header spoofing attacks
- Allows correct client IP extraction for future rate limiting/auth

**Environment Handling:**
- Use TRUSTED_PROXIES environment variable for flexibility
- Default to Fly.io ranges if not specified

### 2. Security Headers Middleware

**Package:** `github.com/gin-contrib/secure`

**Configuration Structure:**
```go
secureConfig := secure.Config{
    AllowedHosts:          config.AllowedHosts,          // From env
    SSLRedirect:           config.Env == "production",
    STSSeconds:            config.HSTSMaxAge,            // 31536000 prod
    STSIncludeSubdomains:  true,
    FrameDeny:             true,                          // X-Frame-Options: DENY
    ContentTypeNosniff:    true,                          // X-Content-Type-Options: nosniff
    BrowserXssFilter:      true,                          // X-XSS-Protection: 1; mode=block
    ReferrerPolicy:        "strict-origin-when-cross-origin",
    ContentSecurityPolicy: buildCSP(config.CSPMode),
}

router.Use(secure.New(secureConfig))
```

**CSP Policy Design:**

*Development (CSP_MODE=relaxed):*
```
default-src 'self';
style-src 'self' 'unsafe-inline';
script-src 'self' 'unsafe-inline';
img-src 'self' data:;
```

*Production (CSP_MODE=strict):*
```
default-src 'self';
style-src 'self' 'unsafe-inline';
script-src 'self';
img-src 'self' https://*.tigris.dev;
object-src 'none';
base-uri 'self';
form-action 'self';
```

**Notes:**
- `'unsafe-inline'` for styles accommodates Hugo themes with inline CSS
- Tigris CDN domain included for future object storage
- Production CSP is stricter but still allows Hugo site functionality
- Future refinement: nonce-based CSP for inline scripts when adding API

**Multi-Domain Support:**
- AllowedHosts from environment variable (comma-separated)
- Middleware validates Host header matches allowed list
- Rejects requests with unknown Host header
- Easy to add memos.alki.me via config change

### 3. Secure Static File Serving

**Current Implementation Issues:**
```go
// INSECURE: Manual file serving with NoRoute
router.NoRoute(func(c *gin.Context) {
    path := "./public" + c.Request.URL.Path  // Path traversal risk!
    c.File(path)
})
```

**New Implementation:**
```go
// SECURE: Use Gin's Static with built-in protection
router.Static("/", "./public")
```

**Benefits of router.Static():**
- Uses Go's `http.FileServer` with path sanitization
- Prevents directory traversal (e.g., `../../etc/passwd`)
- Serves `index.html` for directories (Hugo pretty URLs)
- Correct MIME type detection
- Proper 404 handling
- Battle-tested code path

**Route Precedence:**
- `/health` endpoint registered first (takes precedence)
- Future `/api/v1/*` routes registered before static (takes precedence)
- Static catch-all is last

### 4. Configuration Management

**Stack:**
- `github.com/joho/godotenv` - Load .env files
- `github.com/kelseyhightower/envconfig` - Parse into struct

**Configuration Struct:**
```go
type Config struct {
    // Server
    Env  string `envconfig:"ENV" default:"development"`
    Port string `envconfig:"PORT" default:"8080"`

    // Security
    AllowedHosts   []string `envconfig:"ALLOWED_HOSTS" default:"alkime-memos.fly.dev"`
    TrustedProxies []string `envconfig:"TRUSTED_PROXIES" default:"10.0.0.0/8,172.16.0.0/12"`
    HSTSMaxAge     int      `envconfig:"HSTS_MAX_AGE" default:"31536000"`
    CSPMode        string   `envconfig:"CSP_MODE" default:"relaxed"`

    // Logging
    LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}
```

**Loading Sequence:**
```go
// 1. Load .env file (development convenience)
if err := godotenv.Load(); err != nil {
    if os.IsNotExist(err) {
        log.Println("No .env file found, using environment variables")
    } else {
        log.Printf("Warning: Error loading .env file: %v", err)
    }
}

// 2. Parse environment into config struct
var config Config
if err := envconfig.Process("memos", &config); err != nil {
    log.Fatalf("Failed to process config: %v", err)
}
```

**Error Handling:**
- .env not found: Informational (expected in production)
- .env malformed: Warning (development issue, continue)
- envconfig fails: Fatal (invalid config = unsafe to start)

**Security Notes:**
- `.env` already in `.gitignore` (verified)
- Production uses Fly.io secrets (no .env file)
- Sensitive values (future API keys) follow same pattern
- Configuration validated at startup before server starts

### 5. Structured Logging

**Implementation:** Use Go 1.21+ `log/slog` (standard library)

**Logger Setup:**
```go
logLevel := slog.LevelInfo
if config.Env == "development" {
    logLevel = slog.LevelDebug
}

logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: logLevel,
}))

slog.SetDefault(logger)
```

**Security Event Logging:**
```go
// Startup
logger.Info("Starting Memos server",
    "port", config.Port,
    "env", config.Env,
    "allowed_hosts", config.AllowedHosts,
)

// Configuration issues
logger.Warn("Configuration value outside recommended range",
    "setting", "hsts_max_age",
    "value", config.HSTSMaxAge,
    "recommended_min", 31536000,
)

// Security middleware rejections (via gin-contrib/secure callbacks if needed)
```

**Error Sanitization:**
- Gin's ReleaseMode (set when ENV=production) hides stack traces
- Structured logs to stdout (captured by Fly.io)
- Client-facing errors: generic messages only
- Detailed errors: server logs only

## Implementation Plan

### Phase 1: Dependencies & Configuration
1. Add dependencies: `go get` godotenv, envconfig, gin-contrib/secure
2. Create Config struct in `cmd/server/config.go`
3. Implement configuration loading in `main.go`
4. Add example `.env.example` file

### Phase 2: Core Security
5. Configure SetTrustedProxies with env-driven values
6. Implement security middleware with env-aware config
7. Replace NoRoute with router.Static()
8. Set up structured logging with slog

### Phase 3: Testing & Validation
9. Test with .env file (development)
10. Test with environment variables (production simulation)
11. Verify security headers with curl/browser dev tools
12. Test multi-domain configuration
13. Confirm no security warnings on startup

### Phase 4: Documentation
14. Update CLAUDE.md with new environment variables
15. Create .env.example with all config options
16. Document security decisions in this design doc
17. Add deployment notes for Fly.io secrets

## Testing Strategy

### Manual Testing
```bash
# Development mode with .env
make dev
curl -I http://localhost:8080/

# Verify headers present:
# - X-Frame-Options
# - X-Content-Type-Options
# - Content-Security-Policy
# - (No HSTS in dev mode)

# Production mode simulation
ENV=production PORT=8080 ALLOWED_HOSTS=alkime-memos.fly.dev go run cmd/server/main.go
curl -I http://localhost:8080/

# Verify production headers:
# - All dev headers +
# - Strict-Transport-Security
```

### Security Validation
- Check securityheaders.com (after deployment)
- Verify no Gin warnings on startup
- Test directory traversal: `curl http://localhost:8080/../../../etc/passwd`
- Test host header validation: `curl -H "Host: evil.com" http://localhost:8080/`

## Future Considerations

### API Authentication (Phase 2)
When adding `/api/v1/*` endpoints:
- Add CORS middleware (gin-contrib/cors)
- Configure CORS with ALLOWED_ORIGINS from environment
- Add session/JWT middleware
- Update CSP to allow API domains
- Consider rate limiting (existing foundation supports IP extraction)

### Tigris Object Storage
When integrating Tigris:
- Update CSP `img-src` with Tigris CDN domain
- Configure CORS on Tigris bucket
- Ensure security headers don't interfere with signed URLs

### Monitoring
- Add Prometheus metrics endpoint (gin-contrib/prometheus)
- Track security middleware rejections
- Alert on unusual host header patterns

## Migration Notes

**Breaking Changes:** None - additive only

**Deployment Steps:**
1. Deploy with new dependencies
2. Add environment variables to Fly.io secrets
3. Monitor logs for configuration issues
4. Gradually enforce stricter CSP as issues identified

**Rollback Plan:**
- New config has sensible defaults (can deploy without env vars)
- If issues: revert deployment, analyze logs, adjust config
- No database migrations or data changes involved

## References

- [OWASP Top 10 2021](https://owasp.org/Top10/)
- [Gin Security Best Practices](https://github.com/gin-gonic/gin#security)
- [gin-contrib/secure Documentation](https://github.com/gin-contrib/secure)
- [Fly.io Networking Guide](https://fly.io/docs/networking/)
- [Content Security Policy Reference](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
