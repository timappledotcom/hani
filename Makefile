# Hani Markdown Editor Makefile

BINARY_NAME=hani
MAIN_FILE=main.go
GO_FILES=$(wildcard *.go)

.PHONY: all build test clean run install help

all: build

# Build the binary
build:
	@echo "🏗️  Building Hani..."
	go build -o $(BINARY_NAME) .
	@echo "✅ Build complete!"

# Run tests
test: build
	@echo "🧪 Running tests..."
	./test.sh

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -f $(BINARY_NAME)
	@echo "✅ Clean complete!"

# Run the editor with sample file
run: build
	@echo "🚀 Starting Hani with sample file..."
	./$(BINARY_NAME) sample.md

# Run the editor with no file (new file mode)
new: build
	@echo "🚀 Starting Hani in new file mode..."
	./$(BINARY_NAME)

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "🔍 Running linter..."
	go vet ./...

# Development mode - build and run with sample
dev: build
	@echo "🛠️  Development mode..."
	./$(BINARY_NAME) sample.md

# Show help
help:
	@echo "Hani Markdown Editor - Make Commands"
	@echo "===================================="
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build    - Build the binary"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  run      - Run with sample file"
	@echo "  new      - Run in new file mode"
	@echo "  deps     - Install dependencies"
	@echo "  fmt      - Format code"
	@echo "  lint     - Run linter"
	@echo "  dev      - Development mode"
	@echo "  help     - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build          # Build the binary"
	@echo "  make run            # Run with sample.md"
	@echo "  make new            # Create new file"
	@echo "  make test           # Run all tests"
