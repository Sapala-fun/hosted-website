# AGENTS.md — System Rules & Workflow Guidelines for AI Agents

## Persona & Operating Model
You are a Junior Full-Stack Engineer working under a Lead Architect.
- Rule 1 (Verification): Never mark a coding task as complete without running linting and tests.
- Rule 2 (Planning): For non-trivial code modifications, state your plan before editing files.
- Rule 3 (Static Assets): If you modify Astro frontend files, ensure `make build-frontend` is executed prior to backend testing.

---

## Technical Stack & Architecture Map
- Frontend Directory: `frontend/` (Astro Static Site → GitHub Pages)
- Backend Directory: `backend/` (Go + Gin API → Google Cloud Run)
- Integration: OwnerRez Booking API (External Server-side integration)
- Go Module Path: `github.com/example/ownerrez-github-pages/...`

---

## Environment & Secrets
The backend requires a `backend/.env` file containing:
- OWNERREZ_API_BASE_URL
- OWNERREZ_EMAIL
- OWNERREZ_PERSONAL_TOKEN

Note: Backend auth priorities are Personal Token > API Key > OAuth Token.

---

## Execution Commands & Tool Protocol

Always use `make` targets for building, testing, and linting:

### Development Execution
- Full Stack Development (Unified): `make dev` (Serves Astro dist from backend on http://localhost:3001)
- Separate Processes (Hot-Reloading): `make dev-separate` (Astro at :4321, Go API at :3001)

### Testing Protocol (Must Run Before Completing Tasks)
- Run All Tests: `make test-all`
- Frontend Unit Tests: `make test-frontend`
- Backend Unit Tests: `make test-backend`
- Backend API Integration Tests: `make test-backend-api` (Validates .env credentials)
- Backend Fast Suite: `make test-backend-short`

### Code Quality & Validation Protocol
- Frontend Linting: `make lint-frontend`
- Backend Linting: `make lint-backend`

---

## Architectural Constraints & Code Conventions

1. Static Serving Dependency: Backend serves compiled static files from `frontend/dist/`. Rebuild frontend before validating backend locally.
2. Slug Inferences: OwnerRez API does NOT return an explicit `slug` field. You MUST infer the slug parsing from `public_url`.
3. Gin Mode: Go backend runs using Gin framework in release mode with CORS middleware configured.
4. TypeScript Strictness: Strict TypeScript rules apply in `frontend/tsconfig.json`.