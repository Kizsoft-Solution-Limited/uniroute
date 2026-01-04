.PHONY: dev build test clean fmt lint vet migrate

# Development
dev:
	@echo "Starting development server..."
	@CGO_ENABLED=0 go build -o bin/uniroute-gateway cmd/gateway/main.go
	@./bin/uniroute-gateway

# Build
build:
	@echo "Building binaries..."
	@CGO_ENABLED=0 go build -o bin/uniroute-gateway cmd/gateway/main.go
	@CGO_ENABLED=0 go build -o bin/uniroute cmd/cli/main.go
	@CGO_ENABLED=0 go build -o bin/uniroute-tunnel-server cmd/tunnel-server/main.go

# Test
test:
	@echo "Running tests..."
	@CGO_ENABLED=0 go test -v ./...

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...

# Vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Security scan
security:
	@echo "Running security scan..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	@gosec ./...

# Clean
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run all checks
check: fmt lint vet test security
	@echo "All checks passed!"

