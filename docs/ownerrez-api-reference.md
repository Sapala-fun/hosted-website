# OwnerRez API Reference

> **This file should be updated whenever the OwnerRez API changes.**
> Source documentation: https://api.ownerrez.com/help/v2

## Documentation URLs

| Resource | URL |
|----------|-----|
| API Documentation | https://api.ownerrez.com/help/v2 |
| Machine-readable index | https://api.ownerrez.com/help/v2/index.md |
| LLM-friendly map | https://api.ownerrez.com/llms.txt |

## Base URL

All requests go to `https://api.ownerrez.com`. Paths are versioned under `/v2`.

## Authentication (two options)

- **OAuth 2.0** — send `Authorization: Bearer {token}`
- **HTTP Basic** — API key as username, blank password

## Key Integration Points

| Feature | Endpoint Pattern | Status |
|---------|-----------------|--------|
| Property listing data | `/v2/properties` | Pending integration |
| Availability/calendar | `/v2/availability` | Pending integration |
| Booking/reservation | `/v2/bookings` | Pending integration |
| Guest management | `/v2/guests` | Pending integration |
| Messaging | `/v2/messages` | Pending integration |

## Environment Variables

```bash
export OWNERREZ_API_BASE_URL="https://api.ownerrez.com"
export OWNERREZ_API_KEY="your-api-key"
# or OAuth token:
export OWNERREZ_OAUTH_TOKEN="your-oauth-token"
```

## Backend Structure

| File | Purpose |
|------|---------|
| `backend/main.go` | HTTP server with `/api/health`, `/api/properties`, `/api/book` endpoints |
| `backend/internal/ownerrez/client.go` | OwnerRez client layer for property and booking requests |

## Changelog

| Date | Change |
|------|--------|
| 2026-07-13 | Initial reference captured from https://api.ownerrez.com/help/v2 |

## Update Procedure

When the OwnerRez API changes:

1. Review the updated documentation at https://api.ownerrez.com/help/v2
2. Update this file with new endpoint paths, request/response schemas, and authentication details
3. Update `backend/internal/ownerrez/client.go` to reflect any breaking changes
4. Add a changelog entry above
