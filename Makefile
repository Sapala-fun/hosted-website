.PHONY: dev dev-separate build-backend build-frontend test-backend test-frontend test-all lint-backend lint-frontend lint clean clean-all help

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Run backend + frontend locally (one process)
	@echo "Building Astro frontend..."
	@cd frontend && npm run build
	@echo ""
	@echo "Starting Go backend on http://localhost:3001"
	@cd backend && go env -w GOPROXY=direct && go run .

dev-separate: ## Run backend and frontend as two separate processes (with hot-reload)
	@echo "Terminal 1: cd backend && go run ."
	@echo "Terminal 2: cd frontend && npm run dev"
	@echo ""
	@echo "Visit http://localhost:4321 for the Astro dev server (hot-reload enabled)"
	@echo "API is at http://localhost:3001"

build-backend: ## Build the Go backend binary
	@cd backend && go build -o ../bin/ownerrez-proxy .

build-frontend: ## Build the Astro static site
	@cd frontend && npm run build

test-backend: ## Run backend unit tests
	@cd backend && go test ./...

test-backend-cover: ## Run backend tests with coverage report
	@cd backend && go test -coverprofile=coverage.out ./...
	@cd backend && go tool cover -func=coverage.out

test-backend-short: ## Run backend tests only (no network)
	@cd backend && go test -short ./...

test-frontend: ## Run frontend tests (Vitest)
	@cd frontend && npm test

test-all: ## Run all tests (backend + frontend)
	@echo "Running backend tests..."
	@cd backend && go test ./...
	@echo ""
	@echo "Running frontend tests..."
	@cd frontend && npm test
	@echo ""
	@echo "All tests passed!"

lint-backend: ## Lint Go code (golangci-lint if available, otherwise golint)
	@if command -v golangci-lint &> /dev/null; then \
		cd backend && golangci-lint run ./...; \
	else \
		cd backend && go vet ./...; \
	fi

lint-frontend: ## Lint frontend code (ESLint/Prettier)
	@cd frontend && npm run lint || echo "No lint script found, skipping"

clean: ## Remove build outputs and binaries
	@rm -rf frontend/dist frontend/.astro frontend/src/styles/theme.css bin/ backend/coverage.out
	@echo "Cleaned build artifacts"

clean-all: ## Remove all generated files including dependencies
	@$(MAKE) clean
	@rm -rf frontend/node_modules backend/node_modules
	@echo "Cleaned all artifacts and dependencies"
	@echo "Cleaned build artifacts"
