# Scout MCP Makefile
# Provides common development tasks for building, testing, and running Scout MCP

.PHONY: all build test test-unit test-integration clean fmt vet tidy install help run dev deps check

# Default target
all: build test

# Variables
BINARY_NAME=scout-mcp
BINARY_PATH=./bin/$(BINARY_NAME)
CMD_PATH=./cmd/scout-mcp/main.go
TEST_TIMEOUT=30s

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
BOLD=\033[1m
NC=\033[0m # No Color

## Build Commands

# Build the main binary
build:
	@echo "$(BLUE)🔨 Building $(BINARY_NAME)...$(NC)"
	@mkdir -p bin
	@go build -o $(BINARY_PATH) $(CMD_PATH)
	@echo "$(GREEN)✅ Binary built successfully: $(BINARY_PATH)$(NC)"

# Install binary to GOPATH/bin
install: build
	@echo "$(BLUE)📦 Installing $(BINARY_NAME) to GOPATH/bin...$(NC)"
	@go install $(CMD_PATH)
	@echo "$(GREEN)✅ Installation completed$(NC)"

# Clean build artifacts
clean:
	@echo "$(YELLOW)🧹 Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@go clean
	@echo "$(GREEN)✅ Clean completed$(NC)"

## Testing Commands

# Run all tests
test:
	@echo "$(BLUE)🧪 Running all tests...$(NC)"
	@go test -timeout $(TEST_TIMEOUT) ./test
	@go test -timeout $(TEST_TIMEOUT) ./...
	@echo "$(GREEN)🎉 All tests completed!$(NC)"

# Run tests with coverage
test-cover:
	@echo "$(BLUE)📊 Running tests with coverage...$(NC)"
	@go test -coverprofile=coverage.out ./test
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(NC)"

## Code Quality Commands

# Format all Go code
fmt:
	@echo "$(BLUE)📝 Formatting Go code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

# Vet code for issues
vet:
	@echo "$(BLUE)🔍 Vetting code...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✅ Code vetted$(NC)"

# Tidy dependencies
tidy:
	@echo "$(BLUE)📦 Tidying dependencies...$(NC)"
	@go mod tidy
	@cd test && go mod tidy
	@echo "$(GREEN)✅ Dependencies tidied$(NC)"

# Download dependencies
deps:
	@echo "$(BLUE)📦 Downloading dependencies...$(NC)"
	@go mod download
	@cd test && go mod download
	@echo "$(GREEN)✅ Dependencies downloaded$(NC)"

# Run all code quality checks
check: fmt vet tidy
	@echo "$(GREEN)✅ All code quality checks completed$(NC)"

## Development Commands

# Run the server in development mode
dev: build
	@echo "$(BLUE)🚀 Starting Scout MCP server in development mode...$(NC)"
	@./$(BINARY_PATH)

# Run with specific config path
run: build
	@echo "$(BLUE)🚀 Running Scout MCP server...$(NC)"
	@if [ -z "$(PATH_ARG)" ]; then \
		echo "$(RED)❌ Usage: make run PATH_ARG=/path/to/directory$(NC)"; \
		exit 1; \
	fi
	@./$(BINARY_PATH) $(PATH_ARG)

# Initialize configuration with default path
init: build
	@echo "$(BLUE)⚙️  Initializing Scout MCP configuration...$(NC)"
	@if [ -z "$(PATH_ARG)" ]; then \
		echo "$(RED)❌ Usage: make init PATH_ARG=/path/to/directory$(NC)"; \
		exit 1; \
	fi
	@./$(BINARY_PATH) init $(PATH_ARG)

# Run with only mode (ignore config file)
run-only: build
	@echo "$(BLUE)🚀 Running Scout MCP server in only mode...$(NC)"
	@if [ -z "$(PATH_ARG)" ]; then \
		echo "$(RED)❌ Usage: make run-only PATH_ARG=/path/to/directory$(NC)"; \
		exit 1; \
	fi
	@./$(BINARY_PATH) --only $(PATH_ARG)

## Docker Commands (for future use)

# Build Docker image
docker-build:
	@echo "$(BLUE)🐳 Building Docker image...$(NC)"
	@docker build -t scout-mcp .
	@echo "$(GREEN)✅ Docker image built$(NC)"

## Release Commands

# Create a release build with version info
release:
	@echo "$(BLUE)🚀 Building release version...$(NC)"
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)❌ Usage: make release VERSION=v1.0.0$(NC)"; \
		exit 1; \
	fi
	@mkdir -p bin
	@go build -ldflags="-X main.version=$(VERSION)" -o $(BINARY_PATH) $(CMD_PATH)
	@echo "$(GREEN)✅ Release build completed: $(BINARY_PATH)$(NC)"

## Help

# Show help
help:
	@echo "$(BOLD)Scout MCP Development Commands$(NC)"
	@echo ""
	@echo "$(BOLD)Build Commands:$(NC)"
	@echo "  $(BLUE)build$(NC)         Build the scout-mcp binary"
	@echo "  $(BLUE)install$(NC)       Install binary to GOPATH/bin"
	@echo "  $(BLUE)clean$(NC)         Clean build artifacts"
	@echo ""
	@echo "$(BOLD)Testing Commands:$(NC)"
	@echo "  $(BLUE)test$(NC)          Run all tests (unit + integration)"
	@echo "  $(BLUE)test-unit$(NC)     Run unit tests only"
	@echo "  $(BLUE)test-integration$(NC) Run integration tests only"
	@echo "  $(BLUE)test-coverage$(NC) Run tests with coverage report"
	@echo "  $(BLUE)test-watch$(NC)    Run tests in watch mode (requires entr)"
	@echo ""
	@echo "$(BOLD)Code Quality:$(NC)"
	@echo "  $(BLUE)fmt$(NC)           Format all Go code"
	@echo "  $(BLUE)vet$(NC)           Vet code for issues"
	@echo "  $(BLUE)tidy$(NC)          Tidy dependencies"
	@echo "  $(BLUE)deps$(NC)          Download dependencies"
	@echo "  $(BLUE)check$(NC)         Run all code quality checks"
	@echo ""
	@echo "$(BOLD)Development:$(NC)"
	@echo "  $(BLUE)dev$(NC)           Run server in development mode"
	@echo "  $(BLUE)run$(NC)           Run server with path: make run PATH_ARG=/path"
	@echo "  $(BLUE)init$(NC)          Initialize config: make init PATH_ARG=/path"
	@echo "  $(BLUE)run-only$(NC)      Run in only mode: make run-only PATH_ARG=/path"
	@echo ""
	@echo "$(BOLD)Release:$(NC)"
	@echo "  $(BLUE)release$(NC)       Build release: make release VERSION=v1.0.0"
	@echo "  $(BLUE)docker-build$(NC)  Build Docker image"
	@echo ""
	@echo "$(BOLD)Examples:$(NC)"
	@echo "  make build"
	@echo "  make test"
	@echo "  make run PATH_ARG=~/Projects"
	@echo "  make init PATH_ARG=~/Projects" 
	@echo "  make run-only PATH_ARG=/tmp/safe-dir"
	@echo "  make release VERSION=v1.2.0"