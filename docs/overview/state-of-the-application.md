# State of the Application: Alkime Memos

## Executive Summary

**Alkime Memos** is a production-ready static blog platform with an MVP voice-to-blog automation tool, demonstrating AI-augmented development practices. The platform serves as a personal development blog, combining Hugo static site generation with a security-hardened Go web server and a Voice CLI tool. The application is deployed on Fly.io with a working end-to-end workflow for voice-based content creation.

**Current Status:** Operational production site with MVP voice-to-blog workflow
**Automation Status:** Voice CLI tool provides working MVP workflow from audio recording to blog post with AI-powered content generation

---

## 1. Application Purpose

### Current State (Phase I - MVP)
The application serves as:
- **Personal development blog** at https://memos.alki.me/
- **Learning experiment** in building-in-the-open with AI tools (primarily Claude)
- **DevEx exploration** examining AI's impact on developer productivity
- **Portfolio piece** showcasing production-ready architecture and security practices
- **Voice-to-blog platform** with working MVP workflow: record → transcribe → AI first-draft → AI copy-edit → publish

### Voice CLI Tool (MVP Implementation)
- **Audio recording**: MP3 format with configurable duration/size limits
- **Transcription**: OpenAI Whisper API integration
- **AI content generation**: Anthropic Claude Sonnet 4.5 for drafting and copy-editing
- **Mode system**: Supports both public blog posts ("memos") and personal journal entries
- **Working directory**: ~/Documents/Alkime/Memos (cloud storage compatible)
- **Workflow automation**: Single command for record → transcribe → first-draft flow

### Future Goals (Phase II - Planned)
- **RESTful API**: Backend services under `/api/v1/*` namespace (currently reserved but unimplemented)
- **Media management**: Tigris Object Store integration for audio files and media assets
- **Enhanced observability**: Prometheus metrics and monitoring infrastructure
- **Multi-user support**: Authentication and user management for collaborative workflows

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
- `cmd/server/` - Web server entry point
- `cmd/voice/` - Voice CLI tool entry point
- `internal/config/` - Configuration management
- `internal/logger/` - Structured logging setup
- `internal/server/` - HTTP server and security middleware
- `internal/cli/` - CLI tool packages (audio, transcription, AI, editor)
- `internal/workdir/` - Working directory management

### Voice CLI (Go)
- **CLI Framework**: Kong v1.12.1 (command-line parsing and routing)
- **Audio Capture**: malgo v0.11.24 (cross-platform audio I/O)
- **Audio Encoding**: shine-mp3 v0.1.0 (pure Go MP3 encoder)
- **Transcription**: OpenAI Go SDK v1.12.0 (Whisper API integration)
- **AI Generation**: Anthropic SDK Go v1.18.0 (Claude API integration)

**Voice CLI Package Structure**:
- `internal/cli/audio/` - Audio recording with MP3 encoding
- `internal/cli/audio/device/` - Low-level audio device operations
- `internal/cli/transcription/` - OpenAI Whisper integration
- `internal/cli/ai/` - Anthropic Claude content generation
- `internal/cli/editor/` - Terminal editor integration
- `pkg/collections/` - Utility functions for data manipulation

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

### Production Web Server
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

### Voice CLI Workflow
```
User Command: voice [--mode memos|journal] [--duration 1h] [--max-bytes 256MB]
  ↓
┌─────────────────────────────────────────────────────────────────┐
│ 1. Audio Recording (internal/cli/audio)                         │
│    - Capture from system microphone (malgo)                     │
│    - Encode to MP3 in real-time (shine-mp3)                     │
│    - Enforce duration/size limits with progress display         │
│    - Save to ~/Documents/Alkime/Memos/work/{branch}/            │
│    Output: recording.mp3                                        │
└─────────────────────────────────────────────────────────────────┘
  ↓
┌─────────────────────────────────────────────────────────────────┐
│ 2. Transcription (internal/cli/transcription)                   │
│    - Submit MP3 to OpenAI Whisper API                           │
│    - Receive text transcript                                    │
│    Output: transcript.txt                                       │
└─────────────────────────────────────────────────────────────────┘
  ↓
┌─────────────────────────────────────────────────────────────────┐
│ 3. First Draft (internal/cli/ai)                                │
│    - Send transcript to Anthropic Claude Sonnet 4.5             │
│    - Mode-specific prompts (memos: structured, journal: casual) │
│    - Light cleanup: remove verbal tics, improve clarity         │
│    - Open in $EDITOR for user review/edits                      │
│    Output: first-draft.md (no frontmatter)                      │
└─────────────────────────────────────────────────────────────────┘
  ↓ (User manually runs: voice copy-edit)
┌─────────────────────────────────────────────────────────────────┐
│ 4. Copy Edit (internal/cli/ai)                                  │
│    - Send first-draft to Anthropic Claude Sonnet 4.5            │
│    - Generate Hugo frontmatter (title, date, tags, etc.)        │
│    - Polish grammar, style, markdown formatting                 │
│    - Return structured output via tool use API                  │
│    - Save to content/posts/{YYYY-MM}-{slug}.md                  │
│    - Display changes summary in terminal                        │
│    - Open final post in $EDITOR                                 │
│    Output: Final blog post ready for git commit                 │
└─────────────────────────────────────────────────────────────────┘
```

### Key Architectural Decisions

1. **Hugo as CLI tool**: Uses standard Hugo CLI rather than Go library integration for simpler maintenance and tooling compatibility

2. **`public/` directory strategy**: Generated static files are gitignored and created during Docker build process, not version controlled

3. **Environment-aware security**: Development mode uses relaxed settings (no HSTS, relaxed CSP), production enforces strict security (HSTS enabled, strict CSP)

4. **API namespace reservation**: `/api/v1/*` routes are explicitly reserved for future backend development without conflicting with static file serving

5. **Trusted proxy configuration**: Fly.io-specific configuration (10.0.0.0/8) + local development ranges for proper IP extraction behind reverse proxy

6. **Voice CLI as separate binary**: Standalone CLI tool rather than integrated into web server, enabling independent versioning and distribution

7. **MP3 encoding over WAV**: Pure Go implementation (shine-mp3) for ~5-8x compression, staying well under OpenAI's 25MB API limit

8. **Working directory in cloud storage**: ~/Documents/Alkime/Memos enables iCloud/Dropbox sync for voice recordings and drafts

9. **Two-stage AI workflow**: Separate first-draft and copy-edit commands allow user review and manual edits between AI generations

10. **Mode system for content types**: Distinct prompts and frontmatter for public "memos" vs personal "journal" entries

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
- **Blog Posts**: 6 published posts
  - Development diaries documenting the building process
  - Voice CLI implementation posts
  - DevEx reflections on AI-assisted development
  - Mix of voice-generated and manually written content
- **Static Pages**: 2 pages (README, resume)
- **Homepage**: Custom index with project introduction and pinned posts

### Hugo Configuration & Features
- **Production URL**: https://memos.alki.me/
- **Permalink Structure**:
  - Posts: `/posts/:year/:month/:title/`
  - Pages: `/pages/:contentbasename/`
- **Hugo Features**:
  - Pagination (10 posts per page)
  - RSS feed generation
  - Robots.txt generation
  - Minified HTML output
  - Tag taxonomy
- **Custom Features**:
  - GitHub-style callout blocks (note, tip, warning, important, caution)
  - Image caption shortcode
  - Byline shortcode (generated from frontmatter)
  - Custom frontmatter fields: `voiceBased`, `pinned`, `author`

### Development Workflow
```bash
# Local development (web server)
make dev              # Generate Hugo site + run Go server

# Voice CLI workflow (content creation)
voice                 # Record → transcribe → first-draft (end-to-end)
voice copy-edit       # Polish first-draft → final post
voice devices         # List available audio devices

# Voice CLI individual commands (for debugging)
voice record          # Record audio only
voice transcribe      # Transcribe existing audio
voice first-draft     # Generate first draft from transcript

# Manual content creation (traditional)
hugo new posts/my-post.md

# Code quality
make lint            # Run golangci-lint
make test            # Run test suite
make check           # Run tests + linting (CI simulation)

# Build Voice CLI
make build-voice     # Build bin/voice binary
make install-voice   # Install to $GOPATH/bin

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

**Voice CLI:**
- `OPENAI_API_KEY` - OpenAI API key for Whisper transcription
- `ANTHROPIC_API_KEY` - Anthropic API key for Claude content generation
- `EDITOR` - Terminal editor for reviewing drafts (default: 'open')

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
- **Existing Tests**:
  - Health endpoint validation
  - Audio recorder tests (configuration, limits, progress formatting)
  - Transcription client tests (validation, API integration)
  - AI client tests (slug generation)
  - Collections utility tests
- **CI/CD**: GitHub Actions running tests + linting on all PRs and main branch pushes
- **Linter**: golangci-lint with comprehensive rule set (exhaustruct, goconst, godot, wrapcheck, etc.)

### Code Quality Standards
- **Go Style Guide**: Documented in `docs/guides/go-style-guide.md`
  - Extracted from PR reviews and updated regularly
  - Core guidelines: error wrapping, structured logging, interface usage, SDK types
  - Living document updated as new patterns emerge
- **PR Review Process**: Custom slash command `/address-pr-comments` for systematic feedback incorporation

### Planned Test Expansion
- Security headers validation tests
- Static file serving tests
- Configuration validation tests
- CSP mode behavior tests
- Environment-aware security tests
- End-to-end Voice CLI workflow tests

---

## 8. Gap Analysis: Current vs Planned State

### ✅ Working MVP (Phase I)
- Go web server with security middleware
- Hugo static site generation
- Docker containerization
- Fly.io deployment with health checks
- Environment-based configuration
- Structured logging
- CI/CD pipeline
- Development workflow tooling
- **Voice CLI tool (MVP status)**:
  - Audio recording with MP3 encoding
  - OpenAI Whisper transcription
  - Anthropic Claude AI content generation
  - Mode system (memos vs journal)
  - Recording limits enforcement
  - Editor integration
- **Content generation workflows**:
  - Basic automation: record → transcribe → first-draft
  - Two-stage AI workflow with user review
  - Working directory in cloud storage
- **Content library**: 6 published posts demonstrating the workflow

### 🔄 Partially Implemented
- Testing infrastructure (framework exists, good coverage for Voice CLI, minimal for web server)
- Documentation practices (Go style guide established, extracting learnings from PRs)
- Hugo customizations (callout blocks, shortcodes implemented; more custom features possible)

### ⏳ Planned But Not Implemented (Phase II)
- RESTful API endpoints (`/api/v1/*` namespace reserved)
- Tigris Object Store integration for audio archival
- Prometheus metrics and monitoring
- Multi-user support and authentication
- Expanded test coverage for web server components
- Voice CLI distribution (homebrew, binaries, etc.)

---

## 9. Deployment Status

**Production URL**: https://memos.alki.me/
**Fly.io App**: alkime-memos
**Status**: ✅ Deployed and operational
**Current Branch**: `main`
**Latest Commit**: 83080fc "fix: better signal handling (#31)"
**Health Check**: Passing at `/health` endpoint
**Content**: 6 published blog posts, mix of voice-generated and manually written
**Last Major Feature**: Voice CLI mode system and signal handling improvements

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

9. **Voice CLI working directory** - Audio recordings and drafts stored in `~/Documents/Alkime/Memos/work/{branch}/` for cloud storage sync

10. **Two separate binaries** - `cmd/server/` builds the web server, `cmd/voice/` builds the CLI tool; they're independent

11. **Go Style Guide is living documentation** - Extracted from PR reviews and updated regularly at `docs/guides/go-style-guide.md`

12. **Voice CLI requires API keys** - OpenAI for transcription, Anthropic for content generation; set via environment variables

13. **Mode system affects AI behavior** - "memos" mode generates structured blog posts, "journal" mode creates personal entries

---

## Summary

**Alkime Memos is a production-ready static blog platform with an MVP voice-to-blog automation tool.** It successfully serves content with enterprise-grade security practices while demonstrating modern Go development patterns, AI integration, and infrastructure-as-code practices. The Voice CLI tool provides a working workflow from audio recording to published blog posts, leveraging OpenAI Whisper for transcription and Anthropic Claude for AI-powered content generation.

**Phase I Status**: ✅ MVP Complete
- Production web server with security hardening
- Hugo static site generation with custom features
- Voice CLI tool with working automation workflow
- 6 published blog posts demonstrating the workflow
- Living documentation capturing development practices

**Current State**: Stable, deployed, operational with MVP voice-to-blog automation
**Next Phase**: API development, media management, enhanced observability
**Philosophy**: Building in the open, AI-augmented development, production-quality fundamentals

**Key Achievements**:
- Built working MVP for voice-to-blog workflow
- Developed reusable Go packages for audio, transcription, and AI integration
- Established code quality practices through PR-driven style guide
- Demonstrated practical AI integration in production software
- Maintained security best practices throughout rapid development
