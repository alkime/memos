# Voice-to-Blog Platform

A static site generator platform for converting voice recordings into blog posts, built with Hugo (as a Go library), served by a Go web server, and deployed to Fly.io.

## Project Initialization (For Claude Code)

This section contains instructions for the initial project setup. This is a greenfield project starting from scratch.

### Project Overview

This is a **Phase I implementation** focusing on core static site generation and serving infrastructure. Voice recording and transcription automation are deferred to Phase II.

### Phase I Scope

**In Scope:**
- Hugo-based static site generation (using Hugo as a Go library)
- Go web server with Gin framework for serving static files
- Docker containerization
- Fly.io deployment configuration
- Support for blog posts and static pages (markdown-based)
- Local development environment
- Build process that generates Hugo site during Docker build

**Out of Scope (Phase II):**
- Automated voice recording and transcription
- Tigris Object Store integration for large files
- RESTful API endpoints (namespace reserved)
- Command-line tool for content creation

### Architecture

```
Browser
  ↓
Golang Web Server (Gin)
  ├─ / → Serves Hugo-generated static files from /public
  └─ /api/* → (Reserved for future RESTful API)
       ↓
Hugo Static Site Generator
  ├─ content/ → Markdown files (blog posts + static pages)
  ├─ themes/ → Hugo theme
  └─ public/ → Generated static site (gitignored)
```

**Deployment:** Containerized on Fly.io with Hugo site pre-generated during Docker build.

### Tech Stack

- **Go 1.21+** - Web server and Hugo integration
- **Hugo** - Static site generator (used as Go library)
- **Gin** - Go web framework for HTTP routing
- **Docker** - Containerization
- **Fly.io** - Deployment platform
- **GitHub Actions** - CI/CD (optional for Phase I)

### Project Structure

The repository should be organized as follows:

```
/
├── cmd/
│   └── server/
│       └── main.go          # Go web server entry point
├── content/
│   ├── posts/               # Blog posts (markdown)
│   └── pages/               # Static pages (markdown)
├── themes/
│   └── [theme-name]/        # Hugo theme
├── static/                  # Static assets (images, css, js)
├── public/                  # Generated site (gitignored)
├── Dockerfile               # Multi-stage build
├── fly.toml                 # Fly.io configuration
├── go.mod
├── go.sum
├── .gitignore
├── .dockerignore
└── README.md
```

### Development Workflow

1. **Create content branch:** `git checkout -b post/my-new-post`
2. **Scaffold new post:** `hugo new posts/my-new-post.md` (using Hugo CLI locally)
3. **Edit markdown:** Copy in cleaned-up content from manual transcription process
4. **Local preview:** Run Go server locally to preview changes
5. **Commit and push:** Standard git workflow
6. **Deploy:** Merge to main triggers deployment to Fly.io

### Content Structure

**Blog Posts:**
- Location: `content/posts/`
- Naming: `YYYY-MM-DD-title.md` or `title.md` (Hugo handles dating)
- Front matter: title, date, draft status, tags, categories

**Static Pages:**
- Location: `content/pages/`
- Examples: about.md, contact.md
- Front matter: title, layout

### Build Process

**Docker Multi-Stage Build:**
1. **Stage 1:** Go build environment
   - Copy Go source and Hugo content
   - Run Hugo to generate static site to `public/`
   - Build Go binary

2. **Stage 2:** Minimal runtime image
   - Copy Go binary and `public/` directory
   - Expose port (typically 8080)
   - Run web server

### Go Web Server Requirements

**Primary Functions:**
- Serve static files from `public/` directory
- Handle routing for Hugo's pretty URLs
- Serve index.html for directory requests
- Reserve `/api/*` namespace for future RESTful API

**Example Route Structure:**
```go
router := gin.Default()

// Serve static files from Hugo's public directory
router.Static("/", "./public")

// Future API routes (commented out for Phase I)
// api := router.Group("/api")
// {
//     api.GET("/health", healthCheck)
// }
```

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

Should support hot-reloading or easy rebuild during development.

### Deployment to Fly.io

**Requirements:**
- `fly.toml` configuration file
- Dockerfile optimized for production
- Environment variables (if needed)
- Health check endpoint (optional but recommended)

**Deployment trigger:** Manual for Phase I (automatic via GitHub Actions in future)

### Hugo Theme Selection

**Requirements:**
- Clean, minimal design
- Mobile responsive
- Support for blog post list and individual post pages
- Support for static pages
- Dark mode support (nice to have)

**Options to consider:**
- Start with existing Hugo theme (faster)
- Customize later as needed
- Or build minimal custom theme

**Decision needed:** Choose a starter theme or specify theme requirements.

### Git Ignore

Should ignore:
- `public/` (generated files)
- Go binaries
- `*.log`
- `.DS_Store`
- IDE-specific files
- `node_modules/` (if theme uses npm)

### Environment Configuration

Minimal environment variables for Phase I:
- `PORT` - Server port (default: 8080)
- `ENV` - Environment name (dev/prod)
- `LOG_LEVEL` - Logging verbosity

### Testing Strategy (Future)

For Phase I, manual testing is acceptable:
- Local preview works
- Docker build succeeds
- Deployed site is accessible on Fly.io

Automated tests can be added in Phase II.

### Initial Tasks for Claude Code

1. **Initialize Go module** with appropriate dependencies (Gin, Hugo as library if feasible)
2. **Set up project structure** as outlined above
3. **Create basic Dockerfile** with multi-stage build
4. **Create minimal Go web server** that serves static files from `public/`
5. **Set up Hugo configuration** with basic site metadata
6. **Choose and configure Hugo theme** (or create minimal theme)
7. **Create sample blog post** to test the pipeline
8. **Create fly.toml** for Fly.io deployment
9. **Add appropriate .gitignore** and .dockerignore files
10. **Document any deviations** from this plan in comments or updated README

### Notes and Considerations

**Hugo as Go Library vs CLI:**
- Hugo can be used as a Go library, but it's more common to use the CLI
- For Phase I, using Hugo CLI in Dockerfile is simpler and more maintainable
- Consider: `RUN hugo --minify` in Dockerfile
- Re-evaluate library approach if we need programmatic control in Phase II

**URL Structure:**
- Blog posts: `/posts/title/` or `/YYYY/MM/title/`
- Static pages: `/about/`, `/contact/`
- Decide on permalink structure in Hugo config

**Future API Namespace:**
- Reserve `/api/*` but don't implement yet
- Ensure static file serving doesn't conflict
- Consider `/api/v1/*` for versioning from the start

### Open Questions

1. **Project Name:** What should we call this project? (Update throughout)
2. **Hugo Theme:** Which theme should we start with?
3. **Permalink Structure:** What URL pattern for blog posts?
4. **GitHub Actions:** Set up CI/CD in Phase I or defer?

---

## Standard README Sections (To be filled in after initialization)

### Installation

(Instructions for local setup)

### Usage

(How to run locally, create content, deploy)

### Contributing

(Guidelines for contributions, if applicable)

### License

(Choose appropriate license)