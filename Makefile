.PHONY: help dev build hugo clean docker-build docker-run

# Variables
LOCAL_URL := http://localhost:8080/
BINARY_NAME := server
DOCKER_IMAGE := memos
PORT := 8080

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: build-hugo-dev ## Start Go server for local dev
	@echo "Starting Go web server on port $(PORT)..."
	go run cmd/server/main.go

build-go: ## Generate the Go binary
	@echo "Building Go binary..."
	go build -o $(BINARY_NAME) cmd/server/main.go
	@echo "Binary built: $(BINARY_NAME)"

build-hugo-dev: ## Generate the Hugo site for local dev
	@echo "Generating Hugo site for local development..."
	hugo --baseURL $(LOCAL_URL)

build-hugo: ## Generate Hugo site for production
	@echo "Generating Hugo site for production..."
	hugo

clean: ## Clean generated files
	@echo "Cleaning generated files..."
	rm -rf public/
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container locally
	@echo "Running Docker container on port $(PORT)..."
	docker run -p $(PORT):$(PORT) $(DOCKER_IMAGE)
