# OwnerRez GitHub Pages starter

This repository now contains a starter implementation for a vacation rental website that can be hosted on GitHub Pages while using a small Go-based backend API to connect to OwnerRez-style data and booking flows.

## Structure

- index.html — static landing page for GitHub Pages
- backend/ — Go API service with health, properties, and booking endpoints
- backend/internal/ownerrez/ — OwnerRez client layer for property and booking requests
- .github/workflows/deploy-pages.yml — GitHub Pages deployment workflow

## Local development

1. Start the API server:
   ```bash
   cd backend
   go run .
   ```
2. Serve the frontend from the repository root, for example:
   ```bash
   python3 -m http.server 8000
   ```
3. Visit http://localhost:8000 to view the site.

## OwnerRez integration

To connect the proxy to a real OwnerRez account, set these environment variables before starting the server:

```bash
export OWNERREZ_API_BASE_URL="https://your-ownerrez-instance"
export OWNERREZ_API_KEY="your-api-key"
```

The backend will use them to proxy property and booking requests from the frontend without exposing secrets in the browser.

## Deployment plan

- Host the static frontend on GitHub Pages.
- Deploy the Go API to Render, Fly.io, Railway, Cloud Run, or a VPS.
- Point the frontend to the deployed API URL by setting window.API_BASE_URL in the browser or by changing the frontend script to use the deployed origin.

## Notes

- GitHub Pages is ideal for the public website.
- Booking and OwnerRez API secrets should stay on the backend, not in the browser.
- The current API layer is a starter and should be extended with your real OwnerRez credentials and endpoints once available.
