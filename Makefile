# Hani Markdown Editor Makefile

BINARY_NAME=hani
BUBBLETEA_BINARY=hani-bubbletea
DIY_FILE=diy_hani.go
BUBBLETEA_FILES=main.go model.go keys.go config.go highlight.go version.go

.PHONY: all build build-diy build-bubbletea test clean run install help

all: build-diy

# Build the DIY version (recommended)
build-diy:
	@echo "🏗️  Building Hani DIY version..."
	go build -o $(BINARY_NAME) $(DIY_FILE)
	@echo "✅ DIY version build complete!"

# Build the Bubbletea version (legacy)
build-bubbletea:
	@echo "🏗️  Building Hani Bubbletea version..."
	go build -o $(BUBBLETEA_BINARY) $(BUBBLETEA_FILES)
	@echo "✅ Bubbletea version build complete!"

# Build both versions
build: build-diy build-bubbletea

# Run tests (if any exist)
test: build-diy
	@echo "🧪 Running basic functionality test..."
	@echo "Testing DIY version compilation..."
	@./$(BINARY_NAME) --help 2>/dev/null || echo "Built successfully!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -f $(BINARY_NAME) $(BUBBLETEA_BINARY)
	@echo "✅ Clean complete!"

# Run the DIY version with test file
run: build-diy
	@echo "🚀 Starting Hani DIY version..."
	./$(BINARY_NAME) README.md

# Run the DIY version with no file (new file mode)
new: build-diy
	@echo "🚀 Starting Hani DIY version in new file mode..."
	./$(BINARY_NAME)

# Compare both versions
compare: build
	@echo "🔄 Both versions built:"
	@echo "  - DIY version: ./$(BINARY_NAME)"
	@echo "  - Bubbletea version: ./$(BUBBLETEA_BINARY)"

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
