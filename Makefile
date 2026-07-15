.PHONY: dev dev-separate build-backend build-frontend test-backend clean help

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

test-backend: ## Run backend tests
	@cd backend && go test ./...

clean: ## Remove build outputs and binaries
	@rm -rf frontend/dist frontend/.astro bin/ownerrez-proxy
	@echo "Cleaned build artifacts"
