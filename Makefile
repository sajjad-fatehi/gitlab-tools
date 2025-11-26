.PHONY: build test clean install help

# Build the binary
build:
	@echo "Building gitlab-tools..."
	@go build -o gitlab-tools ./cmd/gitlab-tools

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f gitlab-tools
	@go clean

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Run linter (if golangci-lint is installed)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running linter..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/"; \
	fi

# Install binary to GOPATH/bin
install: build
	@echo "Installing to $(shell go env GOPATH)/bin..."
	@cp gitlab-tools $(shell go env GOPATH)/bin/

# Build for multiple platforms
release:
	@echo "Building releases for multiple platforms..."
	@mkdir -p dist
	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build -o dist/gitlab-tools-linux-amd64 ./cmd/gitlab-tools
	@tar -czf dist/gitlab-tools-linux-amd64.tar.gz -C dist gitlab-tools-linux-amd64
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build -o dist/gitlab-tools-linux-arm64 ./cmd/gitlab-tools
	@tar -czf dist/gitlab-tools-linux-arm64.tar.gz -C dist gitlab-tools-linux-arm64
	@echo "Building for macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 go build -o dist/gitlab-tools-darwin-amd64 ./cmd/gitlab-tools
	@tar -czf dist/gitlab-tools-darwin-amd64.tar.gz -C dist gitlab-tools-darwin-amd64
	@echo "Building for macOS ARM64..."
	@GOOS=darwin GOARCH=arm64 go build -o dist/gitlab-tools-darwin-arm64 ./cmd/gitlab-tools
	@tar -czf dist/gitlab-tools-darwin-arm64.tar.gz -C dist gitlab-tools-darwin-arm64
	@echo "Building for Windows AMD64..."
	@GOOS=windows GOARCH=amd64 go build -o dist/gitlab-tools-windows-amd64.exe ./cmd/gitlab-tools
	@cd dist && zip gitlab-tools-windows-amd64.zip gitlab-tools-windows-amd64.exe
	@echo "Building for Windows ARM64..."
	@GOOS=windows GOARCH=arm64 go build -o dist/gitlab-tools-windows-arm64.exe ./cmd/gitlab-tools
	@cd dist && zip gitlab-tools-windows-arm64.zip gitlab-tools-windows-arm64.exe
	@echo "Generating checksums..."
	@cd dist && sha256sum *.tar.gz *.zip > checksums.txt
	@echo "Release builds completed in dist/"

# Clean release artifacts
clean-release:
	@echo "Cleaning release artifacts..."
	@rm -rf dist

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the gitlab-tools binary"
	@echo "  test          - Run all tests"
	@echo "  clean         - Remove build artifacts"
	@echo "  deps          - Install/update dependencies"
	@echo "  lint          - Run golangci-lint (if installed)"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  release       - Build for all platforms (dist/)"
	@echo "  clean-release - Remove release artifacts"
	@echo "  help          - Show this help message"
