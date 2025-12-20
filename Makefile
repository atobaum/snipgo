.PHONY: build build-cli build-gui dev clean install-deps

# Build CLI only
build-cli:
	@echo "Building CLI..."
	@go build -o bin/snipgo ./cmd/snipgo

# Build GUI (requires wails)
build-gui:
	@echo "Building GUI (requires wails CLI)..."
	@wails build

# Build both
build: build-cli build-gui

# Development mode (GUI)
dev:
	@echo "Starting development server..."
	@wails dev

# Install dependencies
install-deps:
	@echo "Installing Go dependencies..."
	@go mod download
	@echo "Installing frontend dependencies..."
	@cd frontend && pnpm install

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf build/
	@rm -rf frontend/dist/

# Build frontend only
build-frontend:
	@echo "Building frontend..."
	@cd frontend && pnpm run build

