.PHONY: build build-cli build-gui dev clean install-deps lint lint-go lint-ts lint-fix test

# Build CLI only
build-cli:
	@echo "Building CLI..."
	@go build -o bin/snipgo ./cmd/snipgo

# Build GUI (requires wails)
build-gui:
	@echo "Building GUI (requires wails CLI)..."
	@mkdir -p frontend/dist
	@touch frontend/dist/.gitkeep
	@echo "Generating Wails bindings..."
	@wails generate module
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

# Lint Go code
lint-go:
	@echo "Linting Go code..."
	@echo "Running gofmt..."
	@if [ $$(gofmt -s -l . | wc -l) -ne 0 ]; then \
		echo "Error: Code is not formatted. Run 'make lint-fix' to fix."; \
		gofmt -s -d .; \
		exit 1; \
	fi
	@echo "Running go vet..."
	@go vet ./...
	@echo "Running goimports..."
	@if command -v goimports >/dev/null 2>&1; then \
		if [ $$(goimports -l . | grep -v "^$$" | wc -l) -ne 0 ]; then \
			echo "Error: Imports are not formatted. Run 'make lint-fix' to fix."; \
			goimports -d .; \
			exit 1; \
		fi; \
	else \
		echo "goimports not found. Installing..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		if [ $$(goimports -l . | grep -v "^$$" | wc -l) -ne 0 ]; then \
			echo "Error: Imports are not formatted. Run 'make lint-fix' to fix."; \
			goimports -d .; \
			exit 1; \
		fi; \
	fi
	@echo "Go linting passed!"

# Lint TypeScript code
lint-ts:
	@echo "Linting TypeScript code..."
	@cd frontend && pnpm run lint

# Type check TypeScript
type-check:
	@echo "Type checking TypeScript..."
	@cd frontend && pnpm run type-check

# Lint both Go and TypeScript
lint: lint-go lint-ts

# Fix linting issues (auto-fixable)
lint-fix:
	@echo "Fixing linting issues..."
	@cd frontend && pnpm run lint:fix
	@echo "Fixing Go formatting..."
	@gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		goimports -w .; \
	fi
	@echo "Linting issues fixed!"

# Run tests
test:
	@echo "Running Go tests..."
	@go test ./...
	@echo "Running TypeScript tests..."
	@cd frontend && pnpm test


