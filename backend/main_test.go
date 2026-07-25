package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/ownerrez-github-pages/internal/ownerrez"
	"github.com/gin-gonic/gin"
)

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.Default()

	engine.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "ownerrez-proxy-go",
			"timestamp": "now",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestBookHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.Default()
	client := ownerrez.NewClient()

	engine.POST("/api/book", func(c *gin.Context) {
		bookHandler(c, client)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/book", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestBookHandler_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.Default()
	client := ownerrez.NewClient()

	engine.POST("/api/book", func(c *gin.Context) {
		bookHandler(c, client)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/book", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestBookHandler_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.Default()
	client := ownerrez.NewClient()

	engine.POST("/api/book", func(c *gin.Context) {
		bookHandler(c, client)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/book", strings.NewReader(`{"guestName":"Test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
