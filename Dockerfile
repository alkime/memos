# Multi-stage Dockerfile for Memos
# Builds Hugo site during Docker build (public/ is gitignored)

# Stage 1: Build environment with Hugo and Go
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install Hugo
RUN apk add --no-cache hugo

# Copy Hugo content and configuration
COPY content/ ./content/
COPY themes/ ./themes/
COPY static/ ./static/
COPY layouts/ ./layouts/
COPY hugo.yaml ./

# Accept baseURL as build argument (defaults to config value if not provided)
ARG HUGO_BASEURL

# Generate static site (minify is configured in hugo.yaml)
RUN if [ -n "$HUGO_BASEURL" ]; then \
      hugo --baseURL "$HUGO_BASEURL"; \
    else \
      hugo; \
    fi

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server .

# Copy the generated public directory from builder
COPY --from=builder /app/public/ ./public/

# Expose port 8080
EXPOSE 8080

# Set environment to production
ENV ENV=production
ENV GIN_MODE=release

# Run the server
CMD ["./server"]
