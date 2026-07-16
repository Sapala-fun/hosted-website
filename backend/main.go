package main

import (
	"fmt"
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

	// Serve static files from the Astro build output (frontend/dist/)
	cwd, _ := os.Getwd()
	repoRoot := filepath.Dir(cwd)
	frontendDist := filepath.Join(repoRoot, "frontend", "dist")

	// API routes take precedence over static files
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "ownerrez-proxy-go",
			"timestamp": "now",
		})
	})
	r.GET("/api/properties", func(c *gin.Context) {
		propertiesHandler(c, client)
	})
	r.GET("/api/property/:slug", func(c *gin.Context) {
		propertyBySlugHandler(c, client)
	})
	r.GET("/api/property/:slug/details", func(c *gin.Context) {
		propertyDetailsHandler(c, client)
	})
	r.POST("/api/book", func(c *gin.Context) {
		bookHandler(c, client)
	})
	r.GET("/api/availability/:slug", func(c *gin.Context) {
		availabilityHandler(c, client)
	})

	// Serve Astro build output for all other routes
	r.StaticFS("/_astro", http.Dir(filepath.Join(frontendDist, "_astro")))
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(frontendDist, "index.html"))
	})
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if c.Request.Method == http.MethodGet && len(path) > 4 && path[1:5] != "api/" {
			c.File(filepath.Join(frontendDist, "index.html"))
		} else {
			c.AbortWithStatusJSON(404, gin.H{"error": "not_found"})
		}
	})

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

func propertyBySlugHandler(c *gin.Context, client *ownerrez.Client) {
	c.Header("Content-Type", "application/json")
	slug := c.Param("slug")

	property, err := client.GetProperty(slug)
	if err != nil {
		c.JSON(200, gin.H{
			"property": gin.H{
				"id":          slug,
				"name":        slug,
				"slug":        slug,
				"description": "Property data unavailable.",
				"nightlyRate": 0,
				"bedrooms":    0,
				"bathrooms":   0,
			},
			"warning": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"property": property})
}

func propertyDetailsHandler(c *gin.Context, client *ownerrez.Client) {
	c.Header("Content-Type", "application/json")
	slug := c.Param("slug")

	details, err := client.GetPropertyDetails(slug)
	if err != nil {
		c.JSON(200, gin.H{
			"details": gin.H{
				"id":          slug,
				"name":        slug,
				"slug":        slug,
				"description": "Property details unavailable.",
				"nightlyRateMin": 0,
				"nightlyRateMax": 0,
				"bedrooms":     0,
				"bathrooms":    0,
			},
			"warning": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"details": details})
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

func availabilityHandler(c *gin.Context, client *ownerrez.Client) {
	c.Header("Content-Type", "application/json")
	slug := c.Param("slug")

	yearStr := c.Query("year")
	monthStr := c.Query("month")

	if yearStr == "" || monthStr == "" {
		c.JSON(400, gin.H{"error": "year and month query parameters are required"})
		return
	}

	var year, month int
	if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil {
		c.JSON(400, gin.H{"error": "invalid year parameter"})
		return
	}
	if _, err := fmt.Sscanf(monthStr, "%d", &month); err != nil {
		c.JSON(400, gin.H{"error": "invalid month parameter"})
		return
	}

	// First, find the property ID from slug
	properties, err := client.GetProperties()
	if err != nil {
		c.JSON(200, gin.H{
			"dates":  []ownerrez.AvailabilityDate{},
			"slug":   slug,
			"year":   year,
			"month":  month,
			"warning": err.Error(),
		})
		return
	}

	var propertyID int
	var nightlyRate float64
	for _, prop := range properties {
		if s, ok := prop["slug"].(string); ok && s == slug {
			if id, ok := prop["id"].(float64); ok {
				propertyID = int(id)
			}
			if nr, ok := prop["nightlyRate"].(float64); ok {
				nightlyRate = nr
			}
			break
		}
	}

	if propertyID == 0 {
		c.JSON(404, gin.H{"error": "property not found", "slug": slug})
		return
	}

	dates, err := client.GetAvailability(propertyID, year, month)
	if err != nil {
		c.JSON(200, gin.H{
			"dates":   []ownerrez.AvailabilityDate{},
			"slug":    slug,
			"year":    year,
			"month":   month,
			"warning": err.Error(),
		})
		return
	}

	// Attach nightly rate to available dates
	for i := range dates {
		if !dates[i].Blocked {
			dates[i].NightlyRate = nightlyRate
		}
	}

	c.JSON(200, gin.H{
		"dates":  dates,
		"slug":   slug,
		"year":   year,
		"month":  month,
		"propertyID": propertyID,
	})
}
