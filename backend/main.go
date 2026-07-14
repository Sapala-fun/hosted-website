package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/example/ownerrez-github-pages/internal/ownerrez"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {
	client := ownerrez.NewClient()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(corsMiddleware())

	// API routes
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "ok",
				"service":   "ownerrez-proxy-go",
				"timestamp": "now",
			})
		})
		api.GET("/properties", func(c *gin.Context) {
			propertiesHandler(c, client)
		})
		api.POST("/book", func(c *gin.Context) {
			bookHandler(c, client)
		})
	}

	// Serve static files from the repository root (parent of backend/)
	cwd, _ := os.Getwd()
	repoRoot := filepath.Dir(cwd)
	r.StaticFile("/", repoRoot+"/index.html")
	r.StaticFile("/favicon.svg", repoRoot+"/frontend/public/favicon.svg")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("server listening on :%s (frontend: http://localhost:%s)", port, port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func propertiesHandler(c *gin.Context, client *ownerrez.Client) {
	c.Header("Content-Type", "application/json")

	properties, err := client.GetProperties()
	if err != nil {
		c.JSON(200, gin.H{
			"properties": []gin.H{{
				"id":          "sample-property",
				"name":        "Ocean View Retreat",
				"slug":        "ocean-view-retreat",
				"nightlyRate": 210,
				"bedrooms":    2,
				"bathrooms":   2,
				"description": "Fallback property payload served by Go.",
			}},
			"warning": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"properties": properties})
}

func bookHandler(c *gin.Context, client *ownerrez.Client) {
	var payload map[string]string
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	result, err := client.CreateBooking(payload)
	if err != nil {
		c.JSON(200, gin.H{
			"message":  "Booking request received",
			"guestName": payload["guestName"],
			"checkIn":  payload["checkIn"],
			"checkOut": payload["checkOut"],
			"note":     err.Error(),
		})
		return
	}

	c.JSON(200, result)
}
