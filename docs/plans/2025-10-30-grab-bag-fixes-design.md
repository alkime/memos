# Grab Bag Fixes - Design Document

**Date**: 2025-10-30
**Status**: Approved
**Approach**: Four separate PRs for independent review and merge

## Overview

This document outlines four miscellaneous fixes to improve the codebase quality and maintainability. Each task will be implemented as a separate PR to allow independent review, testing, and merge cycles.

## Background

Current state analysis revealed:
- No test coverage exists in the codebase
- PaperMod theme is present but inactive (hugo-bearblog is active)
- Visited links use distinct purple color (#8b6fcb) in hugo-bearblog theme
- Static file serving uses `NoRoute` with wrapped `http.FileServer`

## Goals

1. **Remove unused PaperMod theme** - Clean up git submodule no longer in use
2. **Unify link colors** - Remove visual distinction between visited and unvisited links
3. **Modernize static serving** - Switch to official gin-contrib/static middleware
4. **Establish testing pattern** - Create minimal test coverage using httptest and testify

## Design

### Task 1: Remove PaperMod Theme

**Branch**: `chore/remove-papermod-theme`
**Risk Level**: Low - theme is already inactive

**Implementation Steps**:
1. Deinitialize git submodule: `git submodule deinit -f themes/PaperMod`
2. Remove PaperMod entry from `.gitmodules` file
3. Remove PaperMod configuration from `.git/config` (if present)
4. Delete the theme directory: `git rm -rf themes/PaperMod`
5. Remove commented-out theme line from `hugo.yaml` (line 4: `# theme: "PaperMod"`)
6. Commit all changes with descriptive message

**Verification**:
- Run `git submodule status` - should not list PaperMod
- Run `hugo` to build site successfully with only hugo-bearblog
- Run `make dev` to verify site renders correctly locally
- Check `.gitmodules` file has no PaperMod references

**Files Modified**:
- `.gitmodules`
- `hugo.yaml`
- `themes/PaperMod/` (deleted)

---

### Task 2: Unify Link Colors

**Branch**: `style/unify-link-colors`
**Risk Level**: Low - cosmetic CSS change only

**Current Behavior**:
- Visited links display in purple (#8b6fcb)
- Unvisited links display in blue (light: #3273dc, dark: #8cc2dd)
- Theme defines `--visited-color` variable in `themes/hugo-bearblog/layouts/partials/style.html`

**Desired Behavior**:
- All links (visited and unvisited) use the same color
- Visited links should match unvisited links

**Implementation Strategy**:
Use Hugo's template override system rather than modifying the theme submodule directly. Add CSS override in the existing `layouts/partials/custom_head.html` file.

**Implementation Steps**:
1. Edit `/Users/jamcmdr/dev/alkime/memos/layouts/partials/custom_head.html`
2. Add CSS rule to override visited link styling:
   ```css
   /* Override visited link colors to match unvisited links */
   ul.blog-posts li a:visited {
     color: var(--link-color);
   }
   ```
3. This makes visited links use the default link color instead of `--visited-color`

**Why This Approach**:
- Keeps theme submodule untouched (no conflicts with upstream updates)
- Follows Hugo's standard pattern for theme customization
- Uses existing custom_head.html file already in use for other styles
- Easily reversible if needed

**Verification**:
- Run `make build-hugo-dev` to generate site
- Run `make dev` to test locally
- Visit blog post links in browser
- Confirm visited links look identical to unvisited links
- Check both light and dark modes

**Files Modified**:
- `layouts/partials/custom_head.html`

---

### Task 3: Switch to gin-contrib/static

**Branch**: `refactor/gin-contrib-static`
**Risk Level**: Low-Medium - changes serving mechanism but using well-tested official middleware

**Current Implementation**:
Located in `/Users/jamcmdr/dev/alkime/memos/internal/server/server.go` (lines 70-73):
```go
// Serve static files from Hugo's public directory as fallback
// Using http.FileServer for built-in path traversal protection
// NoRoute only triggers when no explicit routes match (like /health)
s.router.NoRoute(gin.WrapH(http.FileServer(http.Dir("./public"))))
```

**New Implementation**:
Use `gin-contrib/static` middleware for more idiomatic Gin pattern:
```go
s.router.Use(static.Serve("/", static.LocalFile("./public", false)))
```

**Benefits**:
- More idiomatic Gin middleware pattern
- Cleaner API (no manual wrapping of stdlib handler)
- Built-in features like ETag support
- Better integration with Gin's middleware chain
- Improved caching header management

**Implementation Steps**:
1. Add dependency: `go get github.com/gin-contrib/static`
2. Run `go mod tidy` to clean up
3. Edit `internal/server/server.go`:
   - Add import: `"github.com/gin-contrib/static"`
   - Replace `NoRoute` block with middleware approach
   - Keep `/health` endpoint defined explicitly (takes precedence over static serving)
4. Update comments to reflect new approach

**Verification**:
- Run `make build-go` to ensure compilation succeeds
- Run `make dev` to test locally
- Verify `/health` endpoint returns 200 OK
- Verify static files serve correctly from `/`
- Test path traversal protection: `curl http://localhost:8080/../../../etc/passwd` should not leak files
- Test Hugo pretty URLs work correctly
- Test directory index handling

**Files Modified**:
- `go.mod` (new dependency)
- `go.sum` (dependency checksums)
- `internal/server/server.go`

**Rollback Plan**:
If issues arise, easily revert by:
1. Removing gin-contrib/static import
2. Restoring original NoRoute code
3. Running `go mod tidy` to remove unused dependency

---

### Task 4: Add Minimal Test Coverage

**Branch**: `test/initial-test-setup`
**Risk Level**: Low - only adding tests, no production code changes

**Current State**:
- Zero test files exist (glob `**/*_test.go` returns no results)
- No testing frameworks imported beyond stdlib
- No test infrastructure or patterns established

**Goal**:
Establish minimal testing pattern using `httptest` (stdlib) and `testify` (assertions) to serve as foundation for future test expansion.

**Implementation Steps**:

1. **Add testify dependency**:
   ```bash
   go get github.com/stretchr/testify
   go mod tidy
   ```

2. **Create `internal/server/server_test.go`**:
   ```go
   package server

   import (
       "net/http"
       "net/http/httptest"
       "testing"

       "github.com/stretchr/testify/assert"
       "github.com/yourusername/memos/internal/config"
   )

   func TestHealthEndpoint(t *testing.T) {
       // Setup: Create server with test config
       cfg := &config.Config{
           Env:  "test",
           Port: 8080,
       }
       srv := NewServer(cfg)

       // Create test HTTP request
       req := httptest.NewRequest("GET", "/health", nil)
       w := httptest.NewRecorder()

       // Execute request
       srv.router.ServeHTTP(w, req)

       // Assert response using testify
       assert.Equal(t, http.StatusOK, w.Code, "Health endpoint should return 200 OK")
       assert.Contains(t, w.Body.String(), "ok", "Response should contain 'ok'")
   }
   ```

3. **Add `make test` target to Makefile**:
   ```makefile
   .PHONY: test
   test: ## Run tests
       go test ./...
   ```

4. **Update `.golangci.yaml` if needed**:
   - Ensure test files are not flagged with inappropriate linters
   - May need to adjust `exhaustruct` settings for test structs

**Test Scope**:
This PR intentionally keeps scope minimal to:
- Establish the testing pattern and infrastructure
- Demonstrate httptest + testify usage
- Provide a working example for future test additions
- Get CI/testing workflow running

**Future Expansion** (out of scope for this PR):
- Security header tests
- Static file serving tests
- Configuration validation tests
- Error handling tests

**Verification**:
- Run `go mod tidy`
- Run `make test` - tests should pass
- Run `go test -v ./...` - should show test execution details
- Run `go test -cover ./...` - should show coverage metrics
- Run `make lint` - should pass with no test-related errors

**Files Created**:
- `internal/server/server_test.go`

**Files Modified**:
- `go.mod` (testify dependency)
- `go.sum` (dependency checksums)
- `Makefile` (new test target)
- `.golangci.yaml` (potentially, if test adjustments needed)

---

## Execution Order

These PRs can be executed and merged in any order as they are independent. Recommended order by complexity:

1. **PR #1: Remove PaperMod** (simplest, pure cleanup)
2. **PR #2: Unify Link Colors** (simple CSS override)
3. **PR #3: Switch to gin-contrib/static** (trickiest, changes serving mechanism)
4. **PR #4: Add Test Coverage** (independent, establishes new pattern)

## Success Criteria

Each PR must meet these criteria before merge:

- [ ] All existing functionality continues to work
- [ ] `make build-go` succeeds without errors
- [ ] `make lint` passes with no new violations
- [ ] `make dev` runs successfully
- [ ] Manual testing confirms expected behavior
- [ ] Code review approved
- [ ] All verification steps documented in PR description

**Additional for PR #4**:
- [ ] `make test` passes successfully
- [ ] Test coverage metrics visible

## Risk Mitigation

- **Separate PRs**: Each change isolated for easy rollback if issues arise
- **Low-risk first**: Simple changes (theme removal, CSS) before complex (static serving)
- **Verification steps**: Detailed testing checklist for each PR
- **No breaking changes**: All changes backward compatible
- **Manual testing**: Required before merge for each PR

## Future Considerations

Items explicitly out of scope but noted for future:

1. **Expanded test coverage** - security headers, static serving, config validation
2. **CI/CD integration** - run tests automatically on PR
3. **Test coverage reporting** - track coverage metrics over time
4. **Performance testing** - compare gin-contrib/static vs http.FileServer performance
5. **Theme customization** - consider whether to fork hugo-bearblog for deeper customization

## References

- gin-contrib/static: https://github.com/gin-contrib/static
- testify: https://github.com/stretchr/testify
- Hugo template overrides: https://gohugo.io/templates/lookup-order/
- Git submodule documentation: https://git-scm.com/docs/git-submodule
