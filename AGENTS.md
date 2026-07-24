# AGENTS.md — Repository guidance for AI coding agents

Repository overview
-------------------

- **Frontend**: Astro static site (`frontend/`) — GitHub Pages
- **Backend**: Go + Gin API (`backend/`) — Google Cloud Run
- **Integration**: OwnerRez booking API (personal token via backend env)

Local development
-----------------

**Single command (serves both frontend + API):**
```bash
make dev
```
Server runs on `http://localhost:3001`, serves Astro build from `frontend/dist/`.

**Separate processes (development mirroring production):**
```bash
make dev-separate
```
Frontend hot-reload at `http://localhost:4321`, API at `http://localhost:3001`.

Environment variables required in `backend/.env`:
- `OWNERREZ_API_BASE_URL` (required)
- `OWNERREZ_EMAIL` (required)
- `OWNERREZ_PERSONAL_TOKEN` (required)

Testing & validation
--------------------

Run all tests: `make test-all`

Backend-specific:
- Unit tests: `make test-backend`
- API connection tests: `make test-backend-api` (validates `.env` credentials)
- Coverage report: `make test-backend-cover`
- Fast tests only: `make test-backend-short`

Frontend:
- Run tests: `make test-frontend`

Linting
-------

- Backend: `make lint-backend` (golangci-lint or go vet fallback)
- Frontend: `make lint-frontend` (ESLint/Prettier)

Deployment
----------

**Frontend (GitHub Pages):**
- Push to `main` → auto-deploys via `.github/workflows/deploy-pages.yml`

**Backend (Cloud Run):**
- Requires GitHub secrets: `GCLOUD_SERVICE_ACCOUNT_KEY`, `GCLOUD_REGION`, `GCLOUD_PROJECT`, `OWNERREZ_PERSONAL_TOKEN`
- Push to `main` → auto-deploys via `.github/workflows/deploy-backend.yml`

Key gotchas
-----------

1. Backend serves frontend static files from `frontend/dist/`. Run `make build-frontend` before testing backend locally if you've made Astro changes.
2. OwnerRez API returns no explicit `slug` field — backend infers slug from `public_url`.
3. Go module path in imports: `github.com/example/ownerrez-github-pages/...`
4. Backend uses HTTP Basic auth (`OWNERREZ_API_KEY` as base64) or personal token; prioritizes token over key.

Style & conventions
-------------------

- Astro config: strict TypeScript (`frontend/tsconfig.json`)
- Go code: gin framework in `release mode`, CORS middleware enabled for all routes

Help
----

Run `make help` to see all available targets.
