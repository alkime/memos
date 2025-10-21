# Phase I: Core Static Site Generation and Serving

This document outlines the scope, requirements, and implementation checklist for Phase I of the Voice-to-Blog Platform.

## Overview

Phase I focuses on building the foundational infrastructure: a static site generator (Hugo) combined with a Go web server (Gin), containerized with Docker, and deployed to Fly.io.

## Scope

### In Scope

- Hugo-based static site generation (using Hugo CLI, not as Go library)
- Go web server with Gin framework for serving static files
- Docker containerization
- Fly.io deployment configuration
- Support for blog posts and static pages (markdown-based)
- Local development environment
- The `public/` directory is gitignored and built during Docker build

### Out of Scope (Deferred to Phase II)

- Automated voice recording and transcription
- Tigris Object Store integration for large files
- RESTful API endpoints (namespace reserved at `/api/v1/*`)
- Command-line tool for content creation

## Implementation Checklist

When implementing Phase I, ensure:

1. **Go Module Setup**
   - Initialize Go module with appropriate dependencies
   - Add Gin web framework
   - Add any other required dependencies

2. **Project Structure**
   - Create `cmd/server/main.go` for web server entry point
   - Set up `content/posts/` and `content/pages/` directories
   - Set up Hugo theme directory structure
   - Create `static/` directory for assets

3. **Docker Configuration**
   - Create multi-stage Dockerfile
   - Dockerfile should install Hugo and generate `public/` during build
   - Copy Hugo content (content/, themes/, static/, config) to build stage
   - Run `hugo --minify` in build stage to generate static site
   - Optimize for production deployment

4. **Go Web Server**
   - Serve static files from `public/` directory
   - Handle routing for Hugo's pretty URLs
   - Serve index.html for directory requests
   - Reserve `/api/v1/*` namespace (no implementation)

5. **Hugo Configuration**
   - Set up `config.toml` or `config.yaml` with basic site metadata
   - Configure permalink structure
   - Set baseURL appropriately

6. **Hugo Theme**
   - Select and install a Hugo theme, OR
   - Create a minimal custom theme
   - Theme should be clean, minimal, mobile responsive
   - Support blog post lists and individual post pages
   - Dark mode support (nice to have)

7. **Sample Content**
   - Create at least one sample blog post to test the pipeline
   - Create at least one static page (e.g., About)

8. **Fly.io Deployment**
   - Create `fly.toml` configuration file
   - Configure environment variables if needed
   - Set up health check endpoint (optional but recommended)
   - Test deployment manually

9. **Git Configuration**
   - Update `.gitignore` (exclude Go binaries, .env, public/, etc.)
   - **Gitignore `public/`** - it's generated during Docker builds
   - Create `.dockerignore` file

10. **Documentation**
    - Document any deviations from the plan
    - Update README with actual installation/usage instructions
    - Note any open decisions or future improvements

## Hugo Theme Selection

**Requirements:**
- Clean, minimal design
- Mobile responsive
- Support for blog post list and individual post pages
- Support for static pages
- Dark mode support (nice to have)

**Options:**
- Start with existing Hugo theme (faster)
- Customize later as needed
- Or build minimal custom theme from scratch

**Decision needed:** Choose a starter theme or specify custom theme requirements.

## Testing Strategy

For Phase I, manual testing is acceptable:
- Verify local preview works
- Verify Docker build succeeds
- Verify deployed site is accessible on Fly.io
- Test content creation workflow

Automated tests can be added in Phase II.

## Notes and Considerations

### Hugo as Go Library vs CLI

- Hugo can be used as a Go library, but it's more common to use the CLI
- For Phase I, using Hugo CLI is simpler and more maintainable
- Use `hugo --minify` to generate optimized static site
- Re-evaluate library approach if we need programmatic control in Phase II

### Static Site Generation Workflow

Since `public/` is gitignored:
- Run `hugo` locally after content changes for development preview
- Commit only content source files (NOT `public/` directory)
- Docker build installs Hugo and generates `public/` automatically during build process

### URL Structure

Decide on permalink structure in Hugo config:
- Blog posts: `/posts/title/` or `/YYYY/MM/title/`
- Static pages: `/about/`, `/contact/`

### Future API Namespace

- Reserve `/api/v1/*` but don't implement yet
- Ensure static file serving doesn't conflict
- Using `/api/v1/*` for versioning from the start

### Deployment

- Manual deployment for Phase I (run `fly deploy`)
- GitHub Actions CI/CD can be added later

## Open Questions

1. **Hugo Theme:** Which theme should we start with?
2. **Permalink Structure:** What URL pattern for blog posts?
3. **GitHub Actions:** Set up CI/CD in Phase I or defer to Phase II?
4. **Health Check:** Implement health check endpoint or defer?
