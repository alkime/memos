# Multi-stage Dockerfile for Memos
# Since public/ is committed to the repo, we don't need Hugo in the build

# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server/main.go

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server .

# Copy the pre-generated public directory from repo
COPY public/ ./public/

# Expose port 8080
EXPOSE 8080

# Set environment to production
ENV ENV=production
ENV GIN_MODE=release

# Run the server
CMD ["./server"]
