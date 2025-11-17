# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project: Voice-to-Blog Platform

A static site generator platform for converting voice recordings into blog posts, built with Hugo, served by a Go web server, and deployed to Fly.io.

## Architecture

```
Browser
  ↓
Golang Web Server (Gin)
  ├─ / → Serves Hugo-generated static files from /public
  └─ /api/v1/* → (Reserved for future RESTful API)
       ↓
Hugo Static Site Generator
  ├─ content/ → Markdown files (blog posts + static pages)
  ├─ themes/ → Hugo theme
  └─ public/ → Generated static site (gitignored, built during Docker build)
```

**Key Points:**
- Go web server (Gin framework) serves pre-generated Hugo static files
- Hugo site is generated locally for development and during Docker build for production
- The `public/` directory is gitignored - it's not version controlled
- `/api/v1/*` namespace is reserved for future API development
- Deployed as containerized app on Fly.io

## Tech Stack

- **Go** - Web server and CLI tools
- **Hugo** - Static site generator (used via CLI, not as Go library)
- **Gin** - Go web framework
- **shine-mp3** - Pure Go MP3 encoder for audio recording
- **OpenAI Whisper** - Speech-to-text transcription (via API)
- **golangci-lint** - Go linter aggregator for code quality
- **Docker** - Multi-stage build
- **Fly.io** - Deployment platform

## Project Structure

```
/
├── cmd/server/              # Go web server package
│   ├── main.go              # Server entry point
│   ├── config.go            # Configuration management (godotenv + envconfig)
│   └── logging.go           # Structured logging setup
├── content/
│   ├── posts/               # Blog posts (markdown)
│   └── pages/               # Static pages (markdown)
├── themes/[theme-name]/     # Hugo theme (git submodule)
├── static/                  # Static assets
├── public/                  # Generated site (gitignored, built at Docker time)
├── .env.example             # Example environment configuration
├── .env                     # Local environment config (gitignored)
├── Dockerfile               # Multi-stage build
├── fly.toml                 # Fly.io configuration
└── go.mod
```

## Development Commands

A Makefile is provided for common development tasks. Run `make help` to see all available targets.

### Local Development

**Setup:**
1. Copy `.env.example` to `.env` and configure for local development
2. The `.env` file is gitignored and contains local security settings

**Recommended: Using Make**
```bash
# Generate Hugo site with local baseURL and run Go server
make dev
```

**Manual approach (if needed)**
```bash
# Generate site with Hugo CLI (local baseURL)
hugo --baseURL http://localhost:8080/

# Run Go server (loads .env automatically)
go run cmd/server/*.go
```

### Content Creation

```bash
# Create new blog post
hugo new posts/my-new-post.md

# Create new static page
hugo new pages/about.md
```

### Code Quality & Linting

The project uses golangci-lint to enforce code quality standards. A configuration file `.golangci.yaml` is provided in the repository root.

```bash
# Run linter to check code quality
make lint

# Or run golangci-lint directly
golangci-lint run
```

**Linting Standards:**
- All code should pass `make lint` before committing
- Suppression comments (`//nolint:rulename`) are allowed for intentional exceptions (e.g., config structs with optional fields)
- Common enabled linters: goconst, godot, wrapcheck

**Coding Standards:**
@docs/guides/go-style-guide.md

### Testing & Building

```bash
# Build Go binary
make build-go

# Generate Hugo site for local development
make build-hugo-dev

# Generate Hugo site for production
make build-hugo

# Clean generated files
make clean

# Build Docker image
make docker-build

# Run Docker container locally
make docker-run
```

### Deployment

Production configuration is managed in `fly.toml`:

**HTTP Service (`[http_service]`):**
- `force_https = true` - Fly.io proxy handles HTTPS redirect at edge
- `internal_port = 8080` - Go server listens on HTTP internally
- Health check on `/health` endpoint

**Environment Variables (`[env]`):**
- `ENV=production` - Environment mode
- `CSP_MODE=strict` - Content Security Policy mode
- `HSTS_MAX_AGE=31536000` - HSTS max-age (1 year)
- `LOG_LEVEL=info` - Logging verbosity

```bash
# If not created yet...
fly launch

# Deploy on site or config update...
fly deploy
```

## Go Web Server Requirements

The web server must:
- Serve static files from `public/` directory
- Handle routing for Hugo's pretty URLs
- Serve index.html for directory requests
- Reserve `/api/v1/*` namespace for future API development

Example route structure:
```go
router := gin.Default()
router.Static("/", "./public")

// Future API routes:
// api := router.Group("/api/v1")
// { api.GET("/health", healthCheck) }
```

## Docker Build Process

The `public/` directory is gitignored and generated during the Docker build:

1. **Stage 1:** Build environment
   - Install Hugo
   - Copy Hugo content (content/, themes/, static/, config files)
   - Run `hugo` to generate static site
   - Copy Go source
   - Build Go binary

2. **Stage 2:** Runtime image
   - Copy Go binary from build stage
   - Copy generated `public/` directory from build stage
   - Expose port 8080
   - Run web server

## Content Structure

**Blog Posts:**
- Location: `content/posts/`
- Naming: `YYYY-MM-DD-title.md` or `title.md`
- Front matter: title, date, draft, tags, categories

**Static Pages:**
- Location: `content/pages/`
- Front matter: title, layout

## Voice Recording & Transcription

The platform includes a voice CLI tool (`cmd/voice`) for converting voice recordings into blog posts with AI assistance.

**Complete Workflow:**

```bash
# End-to-end: record -> transcribe -> first-draft -> editor
voice

# After editing first draft:
voice copy-edit
```

**Audio Format:**
- **Container:** MP3 (encoded with shine-mp3, pure Go implementation)
- **Sample Rate:** 16,000 Hz (16 kHz) - Whisper's native sample rate
- **Channels:** Mono (1 channel)
- **Compression:** ~5-8x vs WAV (approximately 0.2-0.4 MB per minute)
- **OpenAI API Limit:** 25 MB per request (~1-2 hours of recording possible)

**AI Processing:**
1. **First Draft** (Anthropic Claude Sonnet 4.5):
   - Light cleanup: removes verbal tics (um, like, and)
   - Clarity: rewords for readability while preserving voice
   - Organization: adds section headings, maintains narrative flow
   - Output: Clean markdown (no frontmatter)

2. **Copy Edit** (Anthropic Claude Sonnet 4.5):
   - Polish: grammar, punctuation, style consistency
   - Formatting: proper markdown and Hugo frontmatter
   - Output: Final post in `content/posts/{YYYY-MM}-{slug}.md`

**File Flow:**
```
~/.memos/work/{branch}/
├── recording.mp3       # Step 1: Record
├── transcript.txt      # Step 2: Transcribe (OpenAI Whisper)
└── first-draft.md      # Step 3: First Draft (Anthropic Claude)

content/posts/
└── {YYYY-MM}-{slug}.md # Step 4: Copy Edit (Anthropic Claude)
```

**API Keys:**
- `OPENAI_API_KEY` - For transcription (Whisper)
- `ANTHROPIC_API_KEY` - For AI content generation (Claude)

**Individual Commands:**
- `voice record` - Record audio (auto-transcribes by default)
- `voice transcribe` - Transcribe audio to text
- `voice first-draft` - Generate AI first draft, open in `$EDITOR`
- `voice copy-edit` - Final polish and save to content/posts
- `voice devices` - List available audio devices

See `cmd/voice/README.md` for detailed usage.

## Development Workflow

### Content Changes
1. Create content branch: `git checkout -b post/my-new-post`
2. Scaffold new post: `hugo new posts/my-new-post.md`
3. Edit markdown content
4. Preview locally: `make dev` (calls `make build-hugo-dev` to generate site, then runs Go server)
5. Commit content source files (do NOT commit `public/` directory)
6. Push and create PR
7. Deploy: Merge to main → triggers Fly.io deployment (Docker build generates `public/`)

### Code Changes
1. Create feature branch: `git checkout -b feature/my-feature`
2. Make code changes following the Go Style Guide (see "Coding Standards" section above)
3. Run linter: `make lint` (fix any issues before committing)
4. Test locally: `make dev`
5. Commit changes (do NOT commit `public/` directory or build artifacts)
6. Push and create PR
7. Deploy: Merge to main → triggers Fly.io deployment

## Environment Variables

Configuration is managed via environment variables, loaded from `.env` file in development:

**Server Configuration:**
- `ENV` - Environment mode: `development` or `production`
- `PORT` - Server port (default: 8080)

**Security Configuration:**
- `HSTS_MAX_AGE` - Strict-Transport-Security max-age in seconds (production only, default: 31536000)
- `CSP_MODE` - Content Security Policy mode: `strict`, `relaxed`, or `report-only`

**Logging Configuration:**
- `LOG_LEVEL` - Logging verbosity: `debug`, `info`, `warn`, `error`

See `.env.example` for a complete configuration template.

## Security Features

The server implements OWASP baseline security headers and protection mechanisms:

**Security Headers (Production):**
- `Strict-Transport-Security` (HSTS) - Forces HTTPS for 1 year (production only)
- `X-Frame-Options: DENY` - Prevents clickjacking
- `X-Content-Type-Options: nosniff` - Prevents MIME sniffing
- `X-XSS-Protection: 1; mode=block` - Enables browser XSS protection
- `Referrer-Policy: strict-origin-when-cross-origin` - Limits referrer information
- `Content-Security-Policy` - Configurable CSP (strict/relaxed/report-only modes)

**Protection Mechanisms:**
- **Path Traversal Protection** - Built-in via `http.FileServer`
- **Trusted Platform** - Production uses `gin.PlatformFlyIO` to trust Fly.io proxy headers
- **Structured Logging** - JSON logs with request correlation (slog)

**Environment-Aware Behavior:**
- Development mode (`ENV=development`):
  - No HSTS header (allows HTTP testing)
  - Relaxed CSP by default
  - Debug-level logging available
- Production mode (`ENV=production`):
  - HSTS enabled with configurable max-age
  - Strict CSP recommended
  - Gin release mode
  - No application-level SSL redirect (Fly.io proxy handles HTTPS at edge)

## Important Notes

- **Hugo Usage:** Use Hugo CLI during development, not as Go library. The `public/` directory is gitignored and generated during Docker builds.
- **Static Site Generation:** For local development, use `make dev` which runs `hugo --baseURL http://localhost:8080/` to ensure proper local URL handling. The Docker build process installs Hugo and generates the production site with the production baseURL from `hugo.yaml`.
- **BaseURL Configuration:** The `hugo.yaml` file contains the production baseURL (`https://alkime-memos.fly.dev/`). Local development overrides this using the `--baseURL` flag via `make dev`.
- **API Namespace:** `/api/v1/*` is reserved for future development. Ensure static file serving doesn't conflict.
- **URL Structure:** Permalink structure should be configured in Hugo config (e.g., `/posts/title/` vs `/YYYY/MM/title/`)
