# State of the Application: Alkime Memos

## Executive Summary

**Alkime Memos** is a production-ready static blog platform demonstrating AI-augmented development practices. Currently serving as a personal development blog, it combines Hugo static site generation with a security-hardened Go web server. The application is deployed on Fly.io and represents Phase I completion of a larger vision to automate voice-to-blog workflows.

**Current Status:** Fully operational production site with manual content creation workflow
**Future Vision:** Automated voice memo → LLM processing → blog post publishing pipeline

---

## 1. Application Purpose

### Current State (Phase I - Complete)
The application serves as:
- **Personal development blog** at https://memos.alki.me/
- **Learning experiment** in building-in-the-open with AI tools (primarily Claude)
- **DevEx exploration** examining AI's impact on developer productivity
- **Portfolio piece** showcasing production-ready architecture and security practices
- **Voice-to-blog demonstration** (currently manual workflow: record → transcribe with Claude → edit → publish as markdown)

### Future Goals (Phase II - Planned)
- **Automated content pipeline**: Voice recording → LLM transcription/processing → automated publication
- **RESTful API**: Backend services under `/api/v1/*` namespace (currently reserved but unimplemented)
- **Media management**: Tigris Object Store integration for audio files and media assets
- **CLI tooling**: Command-line tools for streamlined content creation workflow
- **Enhanced observability**: Prometheus metrics and monitoring infrastructure

---

## 2. Technology Stack

### Backend (Go)
- **Runtime**: Go 1.23.0
- **Web Framework**: Gin v1.11.0 (high-performance HTTP framework)
- **Middleware**:
  - `gin-contrib/secure` v1.1.2 - Security headers (HSTS, CSP, X-Frame-Options, etc.)
  - `gin-contrib/static` v1.1.5 - Static file serving
- **Configuration**:
  - `joho/godotenv` v1.5.1 - Environment file loading
  - `kelseyhightower/envconfig` v1.4.0 - Environment variable parsing
- **Logging**: Standard library `log/slog` (structured JSON logging)
- **Testing**: `stretchr/testify` v1.11.1

**Architecture Pattern**: Modular package structure
- `cmd/server/` - Application entry point
- `internal/config/` - Configuration management
- `internal/logger/` - Structured logging setup
- `internal/server/` - HTTP server and security middleware

### Frontend/Content Generation
- **Static Site Generator**: Hugo (CLI-based, not Go library)
- **Theme**: hugo-bearblog (minimalist theme, managed as git submodule)
- **Content Format**: Markdown with YAML frontmatter
- **Template Engine**: Hugo's built-in Go templating

### Infrastructure & Deployment
- **Containerization**: Docker (multi-stage builds)
  - Stage 1: Hugo installation + static site generation + Go binary compilation
  - Stage 2: Alpine-based runtime with compiled binary and generated static files
- **Deployment Platform**: Fly.io
  - Region: San Jose (sjc)
  - Auto-scaling: 0 minimum machines (auto-start/stop)
  - Resources: 256MB RAM, 1 shared CPU
  - HTTPS termination at edge proxy
- **Version Control**: Git with submodules

### Development Tooling
- **Build Automation**: Make (comprehensive Makefile with 15+ targets)
- **Linting**: golangci-lint v2.5.0 (comprehensive configuration with exhaustruct, goconst, godot, wrapcheck)
- **CI/CD**: GitHub Actions (automated testing, linting, code review)
- **Local Development**: Docker Compose support

---

## 3. Architecture Overview

```
Browser
  ↓
Fly.io Proxy (HTTPS termination, force_https=true)
  ↓
Golang Web Server (Gin) - Port 8080
  ├─ Security Middleware
  │  ├─ HSTS (production only, 1 year max-age)
  │  ├─ Content Security Policy (strict/relaxed/report-only modes)
  │  ├─ X-Frame-Options: DENY
  │  ├─ X-Content-Type-Options: nosniff
  │  └─ Referrer-Policy: strict-origin-when-cross-origin
  ├─ /health → Health check endpoint
  ├─ /api/v1/* → Reserved for future API (not implemented)
  └─ /* → Static files from /public directory
       ↓
Hugo Static Site Generator (CLI)
  ├─ content/ → Markdown source files
  ├─ themes/hugo-bearblog/ → Theme (git submodule)
  └─ public/ → Generated static site (gitignored, built at Docker time)
```

### Key Architectural Decisions

1. **Hugo as CLI tool**: Uses standard Hugo CLI rather than Go library integration for simpler maintenance and tooling compatibility

2. **`public/` directory strategy**: Generated static files are gitignored and created during Docker build process, not version controlled

3. **Environment-aware security**: Development mode uses relaxed settings (no HSTS, relaxed CSP), production enforces strict security (HSTS enabled, strict CSP)

4. **API namespace reservation**: `/api/v1/*` routes are explicitly reserved for future backend development without conflicting with static file serving

5. **Trusted proxy configuration**: Fly.io-specific configuration (10.0.0.0/8) + local development ranges for proper IP extraction behind reverse proxy

---

## 4. Security Implementation

The application implements **OWASP baseline security protections**:

### Headers (Production)
- `Strict-Transport-Security` - Forces HTTPS for 1 year (production only, configurable)
- `Content-Security-Policy` - Configurable modes (strict/relaxed/report-only)
- `X-Frame-Options: DENY` - Clickjacking protection
- `X-Content-Type-Options: nosniff` - MIME sniffing protection
- `X-XSS-Protection: 1; mode=block` - Browser XSS filter
- `Referrer-Policy: strict-origin-when-cross-origin` - Privacy protection

### Additional Protections
- Path traversal protection (via `http.FileServer`)
- Trusted proxy configuration for Fly.io deployment
- Environment-based security profiles
- Structured security event logging

### Future Security Considerations (Documented)
When implementing Phase II features:
- CORS middleware for API endpoints
- JWT/session authentication
- Rate limiting (IP extraction already implemented)
- Nonce-based CSP for dynamic content
- Tigris CDN domain whitelisting in CSP

---

## 5. Current Content & Features

### Published Content
- **Blog Posts**: 1 post ("Start With Why" - voice-based, pinned to homepage)
- **Static Pages**: 2 pages (README, resume)
- **Homepage**: Custom index with project introduction

### Hugo Configuration
- **Production URL**: https://memos.alki.me/
- **Permalink Structure**:
  - Posts: `/posts/:year/:month/:title/`
  - Pages: `/pages/:filename/`
- **Features Enabled**:
  - Pagination (10 posts per page)
  - RSS feed generation
  - Robots.txt generation
  - Minified HTML output
  - Tag taxonomy

### Development Workflow
```bash
# Local development (most common)
make dev              # Generate Hugo site + run Go server

# Content creation
hugo new posts/my-post.md

# Code quality
make lint            # Run golangci-lint
make test            # Run test suite
make check           # Run tests + linting (CI simulation)

# Docker workflow
make docker-build    # Build production Docker image
make docker-run      # Run container locally
```

---

## 6. Configuration Management

### Environment Variables
All configuration via environment variables, loaded from `.env` file in development:

**Server:**
- `ENV` - development | production (affects security, logging, Gin mode)
- `PORT` - Server port (default: 8080)

**Security:**
- `HSTS_MAX_AGE` - Seconds for HSTS max-age (default: 31536000 = 1 year, production only)
- `CSP_MODE` - strict | relaxed | report-only (default: strict in prod, relaxed in dev)

**Logging:**
- `LOG_LEVEL` - debug | info | warn | error (default: debug in dev, info in prod)

### Production Configuration (Fly.io)
Set via Fly.io secrets and `fly.toml`:
```toml
[env]
  ENV = "production"
  CSP_MODE = "strict"
  HSTS_MAX_AGE = "31536000"
  LOG_LEVEL = "info"

[http_service]
  internal_port = 8080
  force_https = true
  [http_service.checks.health]
    path = "/health"
    interval = "10s"
```

---

## 7. Testing & Quality Assurance

### Current Test Coverage
- **Test Framework**: testify assertions + standard library `httptest`
- **Existing Tests**: Health endpoint validation (1 test file)
- **CI/CD**: GitHub Actions running tests + linting on all PRs and main branch pushes
- **Linter**: golangci-lint with comprehensive rule set (exhaustruct, goconst, godot, wrapcheck, etc.)

### Planned Test Expansion
Documented in "Grab Bag Fixes" design:
- Security headers validation tests
- Static file serving tests
- Configuration validation tests
- CSP mode behavior tests
- Environment-aware security tests

---

## 8. Gap Analysis: Current vs Planned State

### ✅ Fully Implemented (Phase I)
- Go web server with security middleware
- Hugo static site generation
- Docker containerization
- Fly.io deployment with health checks
- Environment-based configuration
- Structured logging
- CI/CD pipeline
- Development workflow tooling

### 🔄 Partially Implemented
- Testing infrastructure (framework exists, coverage minimal)
- Content library (1 post published, workflow proven)

### ⏳ Planned But Not Implemented (Phase II)
- Voice-to-blog automation pipeline
- RESTful API endpoints (`/api/v1/*` namespace reserved)
- Tigris Object Store integration
- CLI tooling for content creation
- Prometheus metrics and monitoring
- Expanded test coverage

---

## 9. Deployment Status

**Production URL**: https://memos.alki.me/
**Fly.io App**: alkime-memos
**Status**: ✅ Deployed and operational
**Current Branch**: `pinned`
**Latest Commit**: 25b1561 "feat: add pinned to homepage"
**Health Check**: Passing at `/health` endpoint

---

## 10. For Context: What to Know

When working with this codebase, understand:

1. **The `public/` directory is never committed** - It's generated during Docker builds and local `make dev` commands

2. **Hugo is used as a CLI tool** - Don't try to integrate it as a Go library; the Docker build installs and runs Hugo CLI

3. **baseURL handling is environment-aware**:
   - Local dev: `make dev` overrides to `http://localhost:8080/`
   - Production: Uses `hugo.yaml` config (`https://memos.alki.me/`)

4. **Security is environment-aware**:
   - Development mode allows HTTP, has relaxed CSP
   - Production mode enforces HSTS, strict CSP by default

5. **The theme is a git submodule** - Located at `themes/hugo-bearblog/`, managed separately

6. **API namespace is reserved** - `/api/v1/*` routes are protected for future API development

7. **Fly.io handles HTTPS** - The Go server runs HTTP internally; Fly.io proxy handles TLS termination and forced HTTPS redirects

8. **All configuration via environment variables** - No hardcoded config; uses `.env` locally, Fly.io secrets in production

---

## Summary

**Alkime Memos is a production-ready static blog platform in Phase I completion.** It successfully serves content with enterprise-grade security practices, demonstrating modern Go web development patterns and infrastructure-as-code practices. The manual voice-to-blog workflow is proven and operational, setting the stage for Phase II automation features. The codebase is well-structured, documented, and ready for extension with API capabilities and external integrations.

**Current State**: Stable, deployed, operational
**Next Phase**: Automation and API development
**Philosophy**: Building in the open, AI-augmented development, production-quality fundamentals
