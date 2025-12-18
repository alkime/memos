# Architecture: Alkime Memos

## Overview

**Alkime Memos** is a static blog platform with voice-to-blog automation, combining Hugo static site generation with a security-hardened Go web server and a Voice CLI tool. The platform is deployed on Fly.io.

---

## 1. Application Purpose

The application serves as:
- **Personal development blog** at https://memos.alki.me/
- **Learning experiment** in building-in-the-open with AI tools (primarily Claude)
- **DevEx exploration** examining AI's impact on developer productivity
- **Portfolio piece** showcasing production-ready architecture and security practices
- **Voice-to-blog platform** with workflow: record → transcribe → AI first-draft → AI copy-edit → publish

### Voice CLI Tool

- **Audio recording**: MP3 format with configurable duration/size limits
- **Transcription**: OpenAI Whisper API integration
- **AI content generation**: Anthropic Claude Sonnet 4.5 for drafting and copy-editing
- **Mode system**: Supports both public blog posts ("memos") and personal journal entries
- **Working directory**: ~/Documents/Alkime/Memos (cloud storage compatible)
- **Workflow automation**: Single command for record → transcribe → first-draft flow

### Future Goals

- **RESTful API**: Backend services under `/api/v1/*` namespace (currently reserved)
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

**Package Structure**:
- `cmd/server/` - Web server entry point
- `cmd/voice/` - Voice CLI tool entry point
- `internal/platform/config/` - Configuration management
- `internal/platform/logger/` - Structured logging setup
- `internal/platform/server/` - HTTP server and security middleware
- `internal/platform/workdir/` - Working directory management
- `internal/platform/git/` - Git operations (branch detection)
- `internal/platform/keyring/` - Secure credential storage

### Voice CLI (Go)

- **CLI Framework**: Kong v1.12.1 (command-line parsing and routing)
- **TUI Framework**: Bubbletea v1.3.10 (Elm-architecture terminal UI)
  - Bubbles v0.21.0 - Pre-built TUI components
  - Lipgloss v1.1.0 - Terminal styling and layout
- **Audio Capture**: malgo v0.11.24 (cross-platform audio I/O)
- **Audio Encoding**: shine-mp3 v0.1.0 (pure Go MP3 encoder)
- **Transcription**: OpenAI Go SDK v1.12.0 (Whisper API integration)
- **AI Generation**: Anthropic SDK Go v1.18.0 (Claude API integration)

**Voice CLI Package Structure**:
- `internal/audio/` - Audio domain (capture, encoding, recording)
  - `device.go` - Low-level audio device operations
  - `recorder.go` - Audio file recording (PCM buffering, MP3 conversion)
  - `encoder.go` - MP3 encoder configuration
- `internal/content/` - Content transformation domain
  - `transcriber.go` - OpenAI Whisper integration
  - `writer.go` - Anthropic Claude content generation
  - `prompts.go` - AI prompt templates
  - `slug.go` - URL slug generation
- `internal/tui/` - Terminal UI domain
  - `internal/tui/workflow/` - Phase-based workflow states
  - `internal/tui/components/` - Reusable UI components (spinners, phase indicators)
  - `internal/tui/style/` - Lipgloss styling definitions
  - `internal/tui/remotectl/` - UI control interfaces
- `pkg/channels/` - Broadcaster pattern for fan-out channel communication
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

## 3. Architecture Diagrams

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

The Voice CLI uses a Bubbletea-based TUI with a phase-based workflow. The pipeline is resilient to existing outputs—if a phase's output already exists, it can be skipped or resumed.

```
User Command: voice [--mode memos|journal] [--duration 1h] [--max-bytes 256MB]
  ↓
┌─────────────────────────────────────────────────────────────────┐
│                    Bubbletea TUI Application                    │
│                     (internal/tui/model.go)                     │
│                                                                 │
│  Phase indicators show progress through workflow stages         │
│  Press 'q' to quit, Enter to proceed, navigation keys for UI   │
└─────────────────────────────────────────────────────────────────┘
  ↓
┌─────────────────────────────────────────────────────────────────┐
│ Phase 1: Recording (internal/tui/workflow/recording.go)         │
│    - Capture from system microphone (malgo)                     │
│    - Encode to MP3 in real-time (shine-mp3)                     │
│    - Live progress display with duration/size limits            │
│    - Save to ~/Documents/Alkime/Memos/work/{branch}/            │
│    Output: recording.mp3                                        │
│    [Skipped if recording.mp3 exists]                            │
└─────────────────────────────────────────────────────────────────┘
  ↓
┌─────────────────────────────────────────────────────────────────┐
│ Phase 2: Transcribing (internal/tui/workflow/transcribing.go)   │
│    - Submit MP3 to OpenAI Whisper API                           │
│    - Spinner with status updates                                │
│    Output: transcript.txt                                       │
│    [Skipped if transcript.txt exists]                           │
└─────────────────────────────────────────────────────────────────┘
  ↓
┌─────────────────────────────────────────────────────────────────┐
│ Phase 3: View Transcript (internal/tui/workflow/viewtranscript) │
│    - Display transcript for user review                         │
│    - Option to proceed or quit                                  │
└─────────────────────────────────────────────────────────────────┘
  ↓
┌─────────────────────────────────────────────────────────────────┐
│ Phase 4: First Draft (internal/tui/workflow/firstdraft.go)      │
│    - Send transcript to Anthropic Claude Sonnet 4.5             │
│    - Mode-specific prompts (memos: structured, journal: casual) │
│    - Light cleanup: remove verbal tics, improve clarity         │
│    Output: first-draft.md (no frontmatter)                      │
│    [Skipped if first-draft.md exists]                           │
└─────────────────────────────────────────────────────────────────┘
  ↓
┌─────────────────────────────────────────────────────────────────┐
│ Phase 5: Edit Draft (internal/tui/workflow/editdraft.go)        │
│    - Open first-draft.md in $EDITOR for user review/edits       │
│    - User makes manual changes as desired                       │
│    Output: first-draft.md (user-edited)                         │
└─────────────────────────────────────────────────────────────────┘
  ↓
┌─────────────────────────────────────────────────────────────────┐
│ Phase 6: Copy Edit (internal/tui/workflow/copyedit.go)          │
│    - Send first-draft to Anthropic Claude Sonnet 4.5            │
│    - Generate Hugo frontmatter (title, date, tags, etc.)        │
│    - Polish grammar, style, markdown formatting                 │
│    - Return structured output via tool use API                  │
│    - Save to content/posts/{YYYY-MM}-{slug}.md                  │
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

11. **Bubbletea TUI framework**: Elm-architecture terminal UI provides structured state management, composable components, and testable UI logic. Lipgloss handles styling.

12. **Phase-based workflow with resume capability**: Pipeline checks for existing phase outputs (recording.mp3, transcript.txt, first-draft.md) and skips completed phases, enabling interrupted workflows to resume

13. **Broadcaster pattern for concurrent communication**: `pkg/channels` provides fan-out channel utilities for coordinating between audio recording goroutines and the TUI update loop

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

### Future Security Considerations

When implementing future features:
- CORS middleware for API endpoints
- JWT/session authentication
- Rate limiting (IP extraction already implemented)
- Nonce-based CSP for dynamic content
- Tigris CDN domain whitelisting in CSP

---

## 5. Hugo Configuration

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

### Test Infrastructure

- **Test Framework**: testify assertions + standard library `httptest`
- **TUI Testing**: charmbracelet/x/exp/teatest for Bubbletea component testing
- **Test Coverage**:
  - Health endpoint validation
  - Audio recorder tests (configuration, limits, progress formatting)
  - Transcription client tests (validation, API integration)
  - AI client tests (slug generation)
  - Collections utility tests
  - Channel broadcaster tests
- **CI/CD**: GitHub Actions running tests + linting on all PRs and main branch pushes
- **Linter**: golangci-lint with comprehensive rule set (exhaustruct, goconst, godot, wrapcheck, etc.)

### Code Quality Standards

- **Go Style Guide**: Documented in `docs/guides/go-style-guide.md`
  - Extracted from PR reviews and updated regularly
  - Core guidelines: error wrapping, structured logging, interface usage, SDK types
  - Living document updated as new patterns emerge
- **PR Review Process**: Custom slash command `/address-pr-comments` for systematic feedback incorporation

---

## 8. Key Concepts

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

14. **TUI uses Elm architecture** - Bubbletea's Model-Update-View pattern; state changes flow through `Update()`, rendering through `View()`

15. **Pipeline is resumable** - If interrupted, running `voice` again detects existing outputs (recording.mp3, transcript.txt, first-draft.md) and skips completed phases

16. **Broadcaster for concurrent updates** - `pkg/channels.Broadcaster` fans out audio recording progress to multiple subscribers (TUI display, size tracking)
