.PHONY: dev dev-separate build-backend build-frontend test-backend clean help

# Default target
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Start full stack (backend serves both API and frontend)
dev: ## Run backend + frontend locally (one process)
	@cd backend && go env -w GOPROXY=direct && go run .

# Start backend and frontend as separate processes
dev-separate: ## Run backend and frontend as two separate processes
	@echo "Terminal 1: cd backend && go run ."
	@echo "Terminal 2: cd frontend && npm run dev"
	@open -a Terminal.app "$(shell pwd)/scripts/dev-separate.sh" 2>/dev/null || \
		echo "Run the commands manually in two terminals:"

# Build the Go backend
build-backend: ## Build the Go backend binary
	@cd backend && go build -o ../bin/ownerrez-proxy .

# Build the Astro frontend
build-frontend: ## Build the Astro static site
	@cd frontend && npm run build

# Run Go tests
test-backend: ## Run backend tests
	@cd backend && go test ./...

# Clean build artifacts
clean: ## Remove build outputs and binaries
	@rm -rf frontend/dist frontend/.astro bin/ownerrez-proxy
	@echo "Cleaned build artifacts"
