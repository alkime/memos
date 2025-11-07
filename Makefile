.PHONY: help dev build build-voice install-voice hugo clean lint test check docker-build docker-run compose-up compose-down compose-logs

# Variables
LOCAL_URL := http://localhost:8080/
BINARY_NAME := server
VOICE_BINARY := voice
DOCKER_IMAGE := alkime-memos
PORT := 8080

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: build-hugo-dev ## Start Go server for local dev
	@echo "Starting Go web server on port $(PORT)..."
	go run ./cmd/server

build-go: ## Generate the Go binary
	@echo "Building Go binary..."
	go build -o $(BINARY_NAME) ./cmd/server
	@echo "Binary built: $(BINARY_NAME)"

build-voice: ## Build voice CLI binary
	@echo "Building voice CLI..."
	go build -o bin/$(VOICE_BINARY) ./cmd/voice
	@echo "Voice CLI built: bin/$(VOICE_BINARY)"

install-voice: build-voice ## Install voice CLI to $GOPATH/bin
	@echo "Installing voice CLI..."
	@if [ -z "$(GOPATH)" ]; then \
		echo "Error: GOPATH not set"; \
		exit 1; \
	fi
	cp bin/$(VOICE_BINARY) $(GOPATH)/bin/$(VOICE_BINARY)
	@echo "Voice CLI installed to $(GOPATH)/bin/$(VOICE_BINARY)"

build-hugo-dev: clean ## Generate the Hugo site for local dev
	@echo "Generating Hugo site for local development..."
	hugo --baseURL $(LOCAL_URL)

build-hugo: ## Generate Hugo site for production
	@echo "Generating Hugo site for production..."
	hugo

clean: ## Clean generated files
	@echo "Cleaning generated files..."
	rm -rf public/
	rm -f $(BINARY_NAME)
	rm -f bin/$(VOICE_BINARY)
	@echo "Clean complete"

lint: ## Run golangci-lint to check code quality
	@echo "Running golangci-lint..."
	golangci-lint config verify
	golangci-lint run

test: ## Run tests
	@echo "Running tests..."
	go test ./...

check: test lint ## Run tests and linting (CI-ready check)
	@echo "All checks passed!"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container locally
	@echo "Running Docker container on port $(PORT)..."
	docker run -p $(PORT):$(PORT) $(DOCKER_IMAGE)

compose-up: ## Build and run with Docker Compose (production-like testing)
	@echo "Starting Docker Compose (production-like build)..."
	docker compose up --build

compose-down: ## Stop and remove Docker Compose containers
	@echo "Stopping Docker Compose..."
	docker compose down

compose-logs: ## View Docker Compose logs
	@echo "Viewing Docker Compose logs..."
	docker compose logs -f

pr-comments: ## Get any PR comments with the github CLI (optional: make pr-comments PR=123)
	@echo "Getting Github PR comments"
	@./scripts/format_pr.py ${PR}