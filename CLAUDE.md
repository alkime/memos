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

- **Go** - Web server
- **Hugo** - Static site generator (used via CLI, not as Go library)
- **Gin** - Go web framework
- **Docker** - Multi-stage build
- **Fly.io** - Deployment platform

## Project Structure

```
/
├── cmd/server/main.go       # Go web server entry point
├── content/
│   ├── posts/               # Blog posts (markdown)
│   └── pages/               # Static pages (markdown)
├── themes/[theme-name]/     # Hugo theme
├── static/                  # Static assets
├── public/                  # Generated site (gitignored, built at Docker time)
├── Dockerfile               # Multi-stage build
├── fly.toml                 # Fly.io configuration
└── go.mod
```

## Development Commands

A Makefile is provided for common development tasks. Run `make help` to see all available targets.

### Local Development

**Recommended: Using Make**
```bash
# Generate Hugo site with local baseURL and run Go server
make dev
```

**Manual approach (if needed)**
```bash
# Generate site with Hugo CLI (local baseURL)
hugo --baseURL http://localhost:8080/

# Run Go server
go run cmd/server/main.go
```

### Content Creation

```bash
# Create new blog post
hugo new posts/my-new-post.md

# Create new static page
hugo new pages/about.md
```

### Testing & Building

```bash
# Build Go binary
make build

# Generate Hugo site only
make hugo

# Clean generated files
make clean

# Build Docker image
make docker-build

# Run Docker container locally
make docker-run
```

### Deployment

```bash
# If not created yet...
fly launch
```

```bash
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

## Development Workflow

1. Create content branch: `git checkout -b post/my-new-post`
2. Scaffold new post: `hugo new posts/my-new-post.md`
3. Edit markdown content
4. Preview locally: `make dev` (calls `make hugo` to generate site, then runs Go server)
5. Commit content source files (do NOT commit `public/` directory)
6. Push and create PR
7. Deploy: Merge to main → triggers Fly.io deployment (Docker build generates `public/`)

## Environment Variables

- `PORT` - Server port (default: 8080)
- `ENV` - Environment name (dev/prod)
- `LOG_LEVEL` - Logging verbosity

## Important Notes

- **Hugo Usage:** Use Hugo CLI during development, not as Go library. The `public/` directory is gitignored and generated during Docker builds.
- **Static Site Generation:** For local development, use `make dev` which runs `hugo --baseURL http://localhost:8080/` to ensure proper local URL handling. The Docker build process installs Hugo and generates the production site with the production baseURL from `hugo.yaml`.
- **BaseURL Configuration:** The `hugo.yaml` file contains the production baseURL (`https://alkime-memos.fly.dev/`). Local development overrides this using the `--baseURL` flag via `make dev`.
- **API Namespace:** `/api/v1/*` is reserved for future development. Ensure static file serving doesn't conflict.
- **URL Structure:** Permalink structure should be configured in Hugo config (e.g., `/posts/title/` vs `/YYYY/MM/title/`)
