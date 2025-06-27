.PHONY: test test-unit test-integration test-coverage clean build run

# Default test runs only unit tests
test: test-unit

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v ./internal/...

# Run integration tests (requires STABILITY_API_KEY)
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./test/integration/...

# Run all tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run unit tests with coverage for each package
test-coverage-detailed:
	@echo "Running detailed coverage analysis..."
	@mkdir -p coverage
	@go test -v -coverprofile=coverage/api.out ./internal/api
	@go test -v -coverprofile=coverage/storage.out ./internal/storage
	@go test -v -coverprofile=coverage/stability.out ./internal/stability
	@echo "mode: set" > coverage/combined.out
	@tail -n +2 coverage/*.out >> coverage/combined.out
	@go tool cover -html=coverage/combined.out -o coverage/report.html
	@go tool cover -func=coverage/combined.out
	@echo "Detailed coverage report: coverage/report.html"

# Clean build artifacts and test files
clean:
	@echo "Cleaning..."
	rm -f server
	rm -rf coverage/
	rm -f coverage.out coverage.html
	rm -rf uploads/*
	go clean -testcache

# Build the server
build:
	@echo "Building server..."
	go build -o server cmd/server/main.go

# Run the server
run: build
	./server

# Run linting (requires golangci-lint)
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run all checks (fmt, vet, lint, test)
check: fmt vet lint test

# Display help
help:
	@echo "Available targets:"
	@echo "  make test              - Run unit tests"
	@echo "  make test-unit         - Run unit tests (same as 'make test')"
	@echo "  make test-integration  - Run integration tests (requires STABILITY_API_KEY)"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make test-coverage-detailed - Run tests with detailed per-package coverage"
	@echo "  make clean             - Clean build artifacts and test files"
	@echo "  make build             - Build the server binary"
	@echo "  make run               - Build and run the server"
	@echo "  make lint              - Run linters"
	@echo "  make fmt               - Format code"
	@echo "  make vet               - Run go vet"
	@echo "  make check             - Run all checks (fmt, vet, lint, test)"
	@echo "  make help              - Show this help message"