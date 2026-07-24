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
cd backend && ./run-go.sh  # or: go run .
```
Server runs on `http://localhost:3001`, serves Astro build from `frontend/dist/`.

**Separate processes (development mirroring production):**
- Backend: `cd backend && ./run-go.sh` → port 3001
- Frontend: `cd frontend && npm run dev` → port 4321 (calls API at :3001)

Environment variables required in `backend/.env`:
- `OWNERREZ_API_BASE_URL`
- `OWNERREZ_EMAIL`
- `OWNERREZ_PERSONAL_TOKEN`

Testing
-------

- Frontend: `cd frontend && npm run test`
- Backend: `cd backend && go test ./...`

Deployment
----------

**Frontend (GitHub Pages):**
- Push to `main` → auto-deploys via `.github/workflows/deploy-pages.yml`

**Backend (Cloud Run):**
- Requires GitHub secrets: `GCLOUD_SERVICE_ACCOUNT_KEY`, `GCLOUD_REGION`, `GCLOUD_PROJECT`, `OWNERREZ_PERSONAL_TOKEN`
- Push to `main` → auto-deploys via `.github/workflows/deploy-backend.yml`

Key gotchas
-----------

1. Backend serves frontend static files from `frontend/dist/`. Run `cd frontend && npm run build` before testing backend locally if you've made Astro changes.
2. OwnerRez API returns no explicit `slug` field — backend infers slug from `public_url`.
3. Go module path in imports: `github.com/example/ownerrez-github-pages/...`
4. Backend uses HTTP Basic auth (`OWNERREZ_API_KEY` as base64) or personal token; prioritizes token over key.

Style & conventions
-------------------

- Astro config: strict TypeScript (`frontend/tsconfig.json`)
- Go code: gin framework in `release mode`, CORS middleware enabled for all routes
