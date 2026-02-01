# Flight Simulator Makefile

.PHONY: help build run test test-race test-coverage clean fmt vet lint docker-build docker-run

# Variables
BINARY_NAME=simulator
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=cmd/simulator/main.go
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Detect OS for CGO handling
ifeq ($(OS),Windows_NT)
	DETECTED_OS := Windows
else
	DETECTED_OS := $(shell uname -s)
endif

# Default target
help:
	@echo "Flight Simulator - Available targets:"
	@echo ""
	@echo "  make build          - Build the binary"
	@echo "  make run            - Run the simulator"
	@echo "  make test           - Run tests"
	@echo "  make test-race      - Run tests with race detector"
	@echo "  make test-coverage  - Generate test coverage report"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Run go vet"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make demo           - Run interactive demo"
	@echo ""

# Build binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Binary built: $(BINARY_PATH)"

# Build with version info
build-versioned:
	@echo "Building $(BINARY_NAME) with version info..."
	@mkdir -p bin
	go build -ldflags "-X main.version=$(shell git describe --tags --always --dirty)" \
		-o $(BINARY_PATH) $(MAIN_PATH)

# Run the simulator
run:
	@echo "Starting Flight Simulator..."
	go run $(MAIN_PATH)

# Run with custom config
run-config:
	@echo "Starting Flight Simulator with custom config..."
	go run $(MAIN_PATH) -config configs/config.yaml

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with race detector (requires CGO)
test-race:
	@echo "Running tests with race detector..."
	@echo "Note: Race detector requires CGO. Enabling CGO_ENABLED=1..."
	@if command -v gcc >/dev/null 2>&1 || command -v clang >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test -race -v ./...; \
	else \
		echo "Warning: No C compiler found. Race detector requires gcc or clang."; \
		echo "Falling back to regular tests..."; \
		go test -v ./...; \
	fi

# Run tests with race detector multiple times
test-race-stress:
	@echo "Stress testing with race detector (100 iterations)..."
	CGO_ENABLED=1 go test -race -count=100 ./...

# Generate test coverage
test-coverage:
	@echo "Generating test coverage report..."
	go test -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report: $(COVERAGE_HTML)"

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./tests/integration/...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "Clean complete"

# Run interactive demo
demo:
	@echo "Running interactive demo..."
	chmod +x scripts/curl-examples.sh
	./scripts/curl-examples.sh

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Install development tools
dev-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all checks (CI simulation)
ci: fmt vet test
	@echo "All CI checks passed!"

# Run all checks with race detector (requires CGO)
ci-race: fmt vet test-race
	@echo "All CI checks with race detector passed!"

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t flight-simulator:latest .

# Docker run
docker-run:
	@echo "Running in Docker..."
	docker run -p 8080:8080 flight-simulator:latest

# Initialize project (first time setup)
init: deps dev-tools
	@echo "Project initialized!"
	@echo "Run 'make build' to build the binary"
	@echo "Run 'make run' to start the simulator"
