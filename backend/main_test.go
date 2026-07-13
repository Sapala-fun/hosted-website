package main

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/example/ownerrez-github-pages/internal/ownerrez"
)

func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
    rr := httptest.NewRecorder()

    healthHandler(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
    }
}

func TestBookHandler(t *testing.T) {
    body := `{"guestName":"Ada","checkIn":"2026-08-01","checkOut":"2026-08-03"}`
    req := httptest.NewRequest(http.MethodPost, "/api/book", http.NoBody)
    req.Body = http.NoBody
    _ = body
    rr := httptest.NewRecorder()

    client := ownerrez.NewClient()
    bookHandler(rr, req, client)

    if rr.Code != http.StatusBadRequest {
        t.Fatalf("expected bad request for empty body, got %d", rr.Code)
    }
}
