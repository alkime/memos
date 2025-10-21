# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project: Voice-to-Blog Platform

A static site generator platform for converting voice recordings into blog posts, built with Hugo, served by a Go web server, and deployed to Fly.io.

## Plans & Phases

Implementation plans and phase-specific requirements are documented in the `plans/` directory:
- `plans/phase-1.md` - Core static site generation and serving infrastructure
- Future phases will be documented as additional files in the plans directory

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
  └─ public/ → Generated static site (committed to repo)
```

**Key Points:**
- Go web server (Gin framework) serves pre-generated Hugo static files
- Hugo site is generated during development and committed to the repository
- The `public/` directory is NOT gitignored - it's version controlled
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
├── public/                  # Generated site (committed to repo)
├── plans/                   # Implementation plans and phase documentation
├── Dockerfile               # Multi-stage build
├── fly.toml                 # Fly.io configuration
└── go.mod
```

## Development Commands

### Local Development

**Option 1: Direct Go execution**
```bash
# Generate site with Hugo CLI
hugo

# Run Go server
go run cmd/server/main.go
```

**Option 2: Docker Compose (recommended)**
```bash
docker-compose up
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
go build -o server cmd/server/main.go

# Build Docker image
docker build -t memos .

# Run Docker container locally
docker run -p 8080:8080 memos
```

### Deployment

```bash
# Deploy to Fly.io
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

Since `public/` is committed to the repository:

1. **Stage 1:** Build environment
   - Copy Go source
   - Build Go binary

2. **Stage 2:** Runtime image
   - Copy Go binary and `public/` directory from repo
   - Expose port 8080
   - Run web server

Alternatively, if regenerating during build is desired:
- Add Hugo installation to build stage
- Run `hugo --minify` before copying to runtime image

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
4. Generate static site: `hugo` or `hugo --minify`
5. Run Go server locally to preview: `go run cmd/server/main.go`
6. Commit both content and generated `public/` files
7. Push and create PR
8. Deploy: Merge to main → triggers Fly.io deployment

## Environment Variables

- `PORT` - Server port (default: 8080)
- `ENV` - Environment name (dev/prod)
- `LOG_LEVEL` - Logging verbosity

## Important Notes

- **Hugo Usage:** Use Hugo CLI during development, not as Go library. The `public/` directory is committed to the repository.
- **Static Site Generation:** Run `hugo` locally after content changes and commit the generated files. Docker build can skip Hugo generation since files are already in repo.
- **API Namespace:** `/api/v1/*` is reserved for future development. Ensure static file serving doesn't conflict.
- **URL Structure:** Permalink structure should be configured in Hugo config (e.g., `/posts/title/` vs `/YYYY/MM/title/`)
