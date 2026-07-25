.PHONY: dev dev-separate build-backend build-frontend test-backend test-frontend test lint-backend lint-frontend lint clean clean-all help install

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Install pre-commit hooks and dependencies
	@echo "Installing pre-commit..."
	@if command -v pre-commit &> /dev/null; then \
		pre-commit install; \
	else \
		echo " Installing pre-commit via pip..."; \
		pip3 install --user pre-commit black; \
		pre-commit install; \
	fi
	@echo ""
	@echo "Installing frontend dependencies..."
	@if [ -d frontend/node_modules ]; then \
		cd frontend && npm install; \
	else \
		cd frontend && npm install; \
	fi
	@echo ""
	@echo "Install complete! Run 'make dev' to start development."

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
	@./scripts/dev-separate.sh

build-backend: ## Build the Go backend binary
	@cd backend && go build -o ../bin/ownerrez-proxy .

build-frontend: ## Build the Astro static site
	@if [ ! -f frontend/src/styles/theme.css ]; then \
		bash scripts/switch-theme.sh classic; \
	fi
	@cd frontend && npm run build

build: ## Build both frontend and backend
	@echo "Building Astro frontend..."
	@if [ ! -f frontend/src/styles/theme.css ]; then \
		bash scripts/switch-theme.sh classic; \
	fi
	@cd frontend && npm run build
	@echo ""
	@echo "Building Go backend..."
	@mkdir -p bin
	@cd backend && go build -o ../bin/ownerrez-proxy .

test-backend: ## Run backend unit tests
	@cd backend && go test ./...

test-backend-api: ## Run backend API connection tests (validates .env credentials)
	@cd backend && go test -v ./internal/ownerrez

test-backend-cover: ## Run backend tests with coverage report
	@cd backend && go test -coverprofile=coverage.out ./...
	@cd backend && go tool cover -func=coverage.out

test-backend-short: ## Run backend tests only (no network)
	@cd backend && go test -short ./...

test-frontend: ## Run frontend tests (Vitest)
	@cd frontend && npm test

test: ## Run all tests (backend + frontend)
	@make test-backend
	@make test-frontend

lint-backend: ## Lint Go code (golangci-lint if available, otherwise golint)
	@if command -v golangci-lint &> /dev/null; then \
		cd backend && golangci-lint run ./...; \
	else \
		cd backend && go vet ./...; \
	fi
	@echo "Running pre-commit hooks for backend..."
	@if [ -d .git ] && command -v pre-commit &> /dev/null; then \
		pre-commit run --files $$(git ls-files backend/ | tr '\n' ' ') || true; \
	fi

lint-frontend: ## Lint frontend code (ESLint/Prettier)
	@cd frontend && npm run lint || echo "No lint script found, skipping"
	@echo "Running pre-commit hooks for frontend..."
	@if [ -d .git ] && command -v pre-commit &> /dev/null; then \
		pre-commit run --files $$(git ls-files frontend/ | grep -E '\.(js|ts|json)$$' | tr '\n' ' ') || true; \
	fi

lint: ## Run all linters (backend + frontend)
	@echo "Linting backend..."
	@make lint-backend
	@echo ""
	@echo "Linting frontend..."
	@make lint-frontend

clean: ## Remove build outputs and binaries
	@rm -rf frontend/dist frontend/.astro frontend/src/styles/theme.css bin/ backend/coverage.out
	@rm -rf backend/ownerrez-proxy-go backend/server
	@echo "Cleaned build artifacts"

clean-all: ## Remove all generated files including dependencies
	@$(MAKE) clean
	@rm -rf frontend/node_modules backend/node_modules
	@echo "Cleaned all artifacts and dependencies"
