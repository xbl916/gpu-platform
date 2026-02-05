.PHONY: all build run test clean lint docker-build docker-push deploy help

# Build variables
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "")

# Go variables
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)
CGO_ENABLED := 0

all: help

help:
	@echo "GPU Platform - Available Commands:"
	@echo ""
	@echo "  make build          - Build the API server binary"
	@echo "  make run            - Build and run the API server"
	@echo "  make test           - Run unit tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make lint           - Run linter (if installed)"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-push   - Push Docker image to registry"
	@echo "  make deploy        - Deploy to Kubernetes"
	@echo "  make migrate       - Run database migrations"
	@echo ""

build:
	@echo "Building GPU Platform API Server..."
	CGO_ENABLED=$(CGO_ENABLED) go build -ldflags "$(LDFLAGS)" -o bin/api-server ./cmd/api-server
	@echo "Build completed: bin/api-server"

run: build
	@echo "Starting GPU Platform API Server..."
	./bin/api-server

test:
	@echo "Running unit tests..."
	go test -v -race ./... 2>&1 | head -100

test-coverage:
	@echo "Running tests with coverage..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Coverage report generated: coverage.out"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out
	@echo "Clean completed"

lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || echo "Linter not installed (golangci-lint)"

docker-build:
	@echo "Building Docker image..."
	docker build -t gpu-platform/api-server:$(VERSION) -t gpu-platform/api-server:latest -f deployments/docker/Dockerfile.api .
	@echo "Docker image built: gpu-platform/api-server:$(VERSION)"

docker-push:
	@echo "Pushing Docker image..."
	docker push gpu-platform/api-server:$(VERSION)
	docker push gpu-platform/api-server:latest
	@echo "Docker image pushed"

deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -k deployments/k8s/overlays/dev
	@echo "Deployment completed"

migrate:
	@echo "Running database migrations..."
	@echo "Migration command placeholder - implement with your migration tool"

version:
	@echo "GPU Platform Version: $(VERSION)"
	@echo "Git Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

deps:
	@echo "Downloading Go dependencies..."
	go mod download
	@echo "Dependencies downloaded"

deps-update:
	@echo "Updating Go dependencies..."
	go get -u ./...
	go mod tidy
	@echo "Dependencies updated"

deps-verify:
	@echo "Verifying Go dependencies..."
	go mod verify
	@echo "Dependencies verified"
