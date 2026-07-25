package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

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
	// Debug: print environment variables to verify loading
	log.Printf("OWNERREZ_EMAIL=%s", os.Getenv("OWNERREZ_EMAIL"))
	log.Printf("OWNERREZ_API_KEY=%s", os.Getenv("OWNERREZ_API_KEY"))
	log.Printf("OWNERREZ_PERSONAL_TOKEN=%s", os.Getenv("OWNERREZ_PERSONAL_TOKEN"))
	log.Printf("OWNERREZ_API_BASE_URL=%s", os.Getenv("OWNERREZ_API_BASE_URL"))

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
	r.GET("/api/availability/all", func(c *gin.Context) {
		allAvailabilityHandler(c, client)
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
		c.JSON(503, gin.H{
			"error":      "ownerrez API unavailable",
			"warning":    err.Error(),
			"properties": []gin.H{},
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
				"id":             slug,
				"name":           slug,
				"slug":           slug,
				"description":    "Property details unavailable.",
				"nightlyRateMin": 0,
				"nightlyRateMax": 0,
				"bedrooms":       0,
				"bathrooms":      0,
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
			"message":   "Booking request received",
			"guestName": payload["guestName"],
			"checkIn":   payload["checkIn"],
			"checkOut":  payload["checkOut"],
			"note":      err.Error(),
		})
		return
	}

	c.JSON(200, result)
}

func availabilityHandler(c *gin.Context, client *ownerrez.Client) {
	c.Header("Content-Type", "application/json")
	slug := c.Param("slug")

	log.Printf("availabilityHandler - slug from URL: %q", slug)

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

	// Debug: log properties to understand slug matching issue
	log.Printf("availabilityHandler - got %d properties", len(properties))
	if err != nil {
		log.Printf("  Error getting properties: %v", err)
	} else {
		for i, prop := range properties {
			propSlug, ok := prop["slug"].(string)
			log.Printf("  Property %d: id=%v, name=%v, public_url=%v, slug_raw=%q, slug_ok=%v",
				i, prop["id"], prop["name"], prop["public_url"], propSlug, ok)
		}
	}
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

	var propertyID int
	var nightlyRate float64
	log.Printf("availabilityHandler - looking for slug: %q", slug)
	for _, prop := range properties {
		// Check for slug field first (if present in API response)
		if s, ok := prop["slug"].(string); ok && s == slug {
			log.Printf("  MATCHED by slug! id=%v, type(id)=%T, slug=%q", prop["id"], prop["id"], s)
			idVal := prop["id"]
			switch v := idVal.(type) {
			case float64:
				propertyID = int(v)
				log.Printf("    propertyID set to %d from float64", propertyID)
			case string:
				propertyID, _ = strconv.Atoi(v)
				log.Printf("    propertyID set to %d from string", propertyID)
			case int:
				propertyID = v
				log.Printf("    propertyID set to %d from int", propertyID)
			default:
				log.Printf("    WARNING: couldn't convert id to int (type=%T)", v)
			}
			if nr, ok := prop["nightlyRate"].(float64); ok {
				nightlyRate = nr
				log.Printf("    nightlyRate set to %.2f", nightlyRate)
			} else {
				log.Printf("    WARNING: couldn't get nightlyRate")
			}
			break
		}
		// Fallback: check public_url for slug match (OwnerRez API doesn't return explicit slug)
		if s, ok := prop["public_url"].(string); ok {
			// Extract slug from public_url like "https://www.sapala.fun/sapala-all-.../orp5b5eae5x"
			// or just check if the public_url starts with our expected path
			expectedUrl := fmt.Sprintf("https://www.sapala.fun/%s", slug)
			if len(s) > len(expectedUrl) && s[:len(expectedUrl)] == expectedUrl {
				if id, ok := prop["id"].(float64); ok {
					propertyID = int(id)
				}
				if nr, ok := prop["nightlyRate"].(float64); ok {
					nightlyRate = nr
				}
				break
			}
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
		"dates":      dates,
		"slug":       slug,
		"year":       year,
		"month":      month,
		"propertyID": propertyID,
	})
}

func allAvailabilityHandler(c *gin.Context, client *ownerrez.Client) {
	c.Header("Content-Type", "application/json")

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

	properties, err := client.GetProperties()
	if err != nil {
		c.JSON(200, gin.H{
			"dates":   []ownerrez.AvailabilityDate{},
			"warning": err.Error(),
		})
		return
	}

	var propertyID int
	for _, prop := range properties {
		if id, ok := prop["id"].(float64); ok {
			propertyID = int(id)
			break
		}
		if id, ok := prop["id"].(int); ok {
			propertyID = id
			break
		}
	}

	if propertyID == 0 {
		c.JSON(200, gin.H{
			"dates":   []ownerrez.AvailabilityDate{},
			"warning": "No properties found",
		})
		return
	}

	dates, err := client.GetAvailability(propertyID, year, month)
	if err != nil {
		c.JSON(200, gin.H{
			"dates":   []ownerrez.AvailabilityDate{},
			"warning": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"dates":      dates,
		"year":       year,
		"month":      month,
		"propertyID": propertyID,
	})
}
