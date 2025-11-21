# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project: Voice-to-Blog Platform

A static site generator platform for converting voice recordings into blog posts, built with Hugo, served by a Go web server, and deployed to Fly.io.

**Architecture Reference:** @docs/overview/architecture.md

**Coding Standards:** @docs/guides/go-style-guide.md

## Quick Reference

```
/
├── cmd/server/              # Go web server
├── cmd/voice/               # Voice CLI tool
├── content/posts/           # Blog posts (markdown)
├── content/pages/           # Static pages (markdown)
├── themes/hugo-bearblog/    # Hugo theme (git submodule)
├── public/                  # Generated site (gitignored)
└── internal/                # Go packages
```

## Development Commands

Run `make help` to see all available targets.

### Local Development

```bash
# Setup: copy .env.example to .env and configure
cp .env.example .env

# Generate Hugo site and run Go server
make dev
```

### Voice CLI

```bash
# End-to-end workflow
voice                 # Record → transcribe → first-draft → editor
voice copy-edit       # Polish first-draft → final post

# Individual commands
voice record          # Record audio only
voice transcribe      # Transcribe existing audio
voice first-draft     # Generate AI first draft
voice devices         # List audio devices
```

### Content Creation

```bash
hugo new posts/my-new-post.md    # New blog post
hugo new pages/about.md          # New static page
```

### Code Quality

```bash
make lint             # Run golangci-lint (required before commits)
make test             # Run test suite
make check            # Run tests + linting (CI simulation)
```

### Building

```bash
make build-go         # Build Go server binary
make build-voice      # Build Voice CLI binary
make build-hugo-dev   # Generate Hugo site (local baseURL)
make build-hugo       # Generate Hugo site (production baseURL)
make docker-build     # Build production Docker image
make docker-run       # Run Docker container locally
make clean            # Clean generated files
```

### Deployment

```bash
fly launch            # Initial setup (if not created)
fly deploy            # Deploy to Fly.io
```

## Development Workflow

### Content Changes

1. Create branch: `git checkout -b post/my-new-post`
2. Scaffold post: `hugo new posts/my-new-post.md`
3. Edit markdown content
4. Preview: `make dev`
5. Commit source files (NOT `public/` directory)
6. Push and create PR
7. Merge to main → Fly.io deployment

### Code Changes

1. Create branch: `git checkout -b feature/my-feature`
2. Make changes following the Go Style Guide
3. Run linter: `make lint`
4. Test locally: `make dev`
5. Commit changes (NOT `public/` or build artifacts)
6. Push and create PR
7. Merge to main → Fly.io deployment

## Go Web Server

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

The `public/` directory is gitignored and generated during Docker build:

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

## Important Notes

- **Hugo Usage:** Use Hugo CLI during development, not as Go library. The `public/` directory is gitignored and generated during Docker builds.
- **Static Site Generation:** `make dev` runs `hugo --baseURL http://localhost:8080/` for local development. Docker build uses production baseURL from `hugo.yaml`.
- **BaseURL Configuration:** `hugo.yaml` contains production baseURL (`https://memos.alki.me/`). Local dev overrides via `--baseURL` flag.
- **API Namespace:** `/api/v1/*` is reserved for future development. Ensure static file serving doesn't conflict.
- **URL Structure:** Permalink structure is configured in Hugo config (`/posts/:year/:month/:title/`)
