# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=url-shortener
BINARY_UNIX=$(BINARY_NAME)_unix

# Build parameters
MAIN_PATH=./cmd/server/main.go
BUILD_PATH=./build

# Directories
BIN_DIR=$(BUILD_PATH)/bin
SRC_DIR=.

.PHONY: all test clean install uninstall run docker-build docker-run

# Build the project
all: clean build

# Build binary
build: 
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed. Binary located at $(BIN_DIR)/$(BINARY_NAME)"

# Run the application
run:
	@echo "Starting $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

# Test the application
test:
	@echo "Running tests..."
	$(GOTEST) ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -f $(BIN_DIR)/$(BINARY_NAME)
	rm -rf $(BUILD_PATH)

# Install the binary to GOBIN or GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOGET) -u
	$(GOBUILD) -o $(GOBIN)/$(BINARY_NAME) $(MAIN_PATH)

# Run the application with specific environment
dev:
	@echo "Starting development server..."
	PORT=8080 DEBUG=true DATABASE_URL=./dev.db BASE_URL=http://localhost:8080 $(GOCMD) run $(MAIN_PATH)

# Build for production
prod:
	@echo "Building production binary..."
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(BIN_DIR)/$(BINARY_UNIX) $(MAIN_PATH)

# Show application information
info:
	@echo "Project: URL Shortener Service"
	@echo "Binary: $(BINARY_NAME)"
	@echo "Main Path: $(MAIN_PATH)"
	@echo "Build Path: $(BUILD_PATH)"

# Help target
help:
	@echo "Usage:"
	@echo "  make build     - Build the application"
	@echo "  make run       - Run the application"
	@echo "  make test      - Run tests"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make dev       - Run in development mode"
	@echo "  make prod      - Build for production"
	@echo "  make info      - Show application information"
	@echo "  make help      - Show this help message"