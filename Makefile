.PHONY: all build test lint clean coverage benchmark help

# Variables
BINARY_NAME=jimi-decoder
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

all: lint test build

## build: Build all binaries
build:
	@echo "Building binaries..."
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/decoder-cli ./cmd/decoder-cli
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/tcp-server ./cmd/tcp-server
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/simulator ./cmd/simulator
	@echo "Build complete!"

## test: Run all tests
test:
	@echo "Running tests..."
	@$(GO) test -v -race -timeout 30s ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@golangci-lint run ./...

## benchmark: Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	@$(GO) test -bench=. -benchmem -run=^$$ ./test/benchmark/

## fuzz: Run fuzz tests (requires Go 1.18+)
fuzz:
	@echo "Running fuzz tests..."
	@$(GO) test -fuzz=. -fuzztime=30s ./test/fuzz/

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf dist/
	@rm -f coverage.out coverage.html
	@$(GO) clean
	@echo "Clean complete!"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy

## update-deps: Update dependencies
update-deps:
	@echo "Updating dependencies..."
	@$(GO) get -u ./...
	@$(GO) mod tidy

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@$(GO) vet ./...

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed!"

## help: Show this help message
help:
	@echo "Available targets:"
	@echo "  make build         - Build all binaries"
	@echo "  make test          - Run all tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make lint          - Run linter"
	@echo "  make benchmark     - Run benchmarks"
	@echo "  make fuzz          - Run fuzz tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make deps          - Download dependencies"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Run go vet"
	@echo "  make install-tools - Install development tools"
	@echo "  make help          - Show this help message"
