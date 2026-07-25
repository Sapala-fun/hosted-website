package ownerrez

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadEnvFile() {
	// Try to load .env file from parent directory (hosted-website/backend/)
	envPath := filepath.Join("..", "..", ".env")
	file, err := os.Open(envPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		os.Setenv(key, value)
	}
}

func TestGetProperties(t *testing.T) {
	// Load env from parent directory's .env
	loadEnvFile()

	client := NewClient()

	// Verify credentials are configured
	if client.BaseURL == "" {
		t.Fatal("OWNERREZ_API_BASE_URL is not configured")
	}

	authHeader := client.authHeader()
	if authHeader == "" {
		t.Fatal("Authentication header is empty - credentials may be misconfigured")
	}

	fmt.Printf("Testing with BaseURL: %s\n", client.BaseURL)
	fmt.Printf("Auth Header (truncated): %s...\n", authHeader[:30])

	properties, err := client.GetProperties()
	if err != nil {
		t.Fatalf("GetProperties() failed: %v", err)
	}

	fmt.Printf("Successfully retrieved %d properties\n", len(properties))

	if len(properties) == 0 {
		t.Log("Warning: No properties returned from API")
	} else {
		for i, prop := range properties {
			name := prop["name"]
			publicURL := prop["public_url"]
			fmt.Printf("  [%d] %v - %v\n", i+1, name, publicURL)
		}
	}
}

func TestGetPropertyDetails(t *testing.T) {
	loadEnvFile()
	client := NewClient()

	if client.BaseURL == "" || client.authHeader() == "" {
		t.Fatal("Credentials not configured")
	}

	// First get properties to find a valid slug
	properties, err := client.GetProperties()
	if err != nil {
		t.Fatalf("GetProperties() failed: %v", err)
	}

	if len(properties) == 0 {
		t.Fatal("No properties found to test with")
	}

	// Use the first property's public_url slug (extract from URL)
	slug := ""
	for _, prop := range properties {
		if publicURL, ok := prop["public_url"].(string); ok && publicURL != "" {
			// Extract slug from URL like https://www.sapala.fun/sapala-all-exclusive-oceanfront-villa-with-private-pool-orp5b5eae5x
			parts := strings.Split(strings.TrimSuffix(publicURL, "/"), "/")
			if len(parts) > 0 {
				slug = parts[len(parts)-1]
				break
			}
		}
	}

	if slug == "" {
		t.Fatal("Could not extract property slug from properties")
	}

	fmt.Printf("Testing GetPropertyDetails for slug: %s\n", slug)

	details, err := client.GetPropertyDetails(slug)
	if err != nil {
		t.Fatalf("GetPropertyDetails(%q) failed: %v", slug, err)
	}

	fmt.Printf("\nProperty Details for '%s':\n", details.ID)
	fmt.Printf("  ID: %s\n", details.ID)
	fmt.Printf("  Name: %s\n", details.Name)
	fmt.Printf("  Bedrooms: %d\n", details.Bedrooms)
	fmt.Printf("  Bathrooms: %d\n", details.Bathrooms)
	fmt.Printf("  Sleeps: %s\n", details.Sleeps)
	fmt.Printf("  Nightly Rate Min: $%.2f\n", details.NightlyRateMin)
	fmt.Printf("  Nightly Rate Max: $%.2f\n", details.NightlyRateMax)
	fmt.Printf("  Location: %s\n", details.Location)
	fmt.Printf("  Booking URL: %s\n", details.BookingURL)
	fmt.Printf("  Description preview: %.100s...\n", details.Description)
}

func TestClientAuthHeader(t *testing.T) {
	loadEnvFile()
	client := NewClient()

	header := client.authHeader()

	if header == "" {
		t.Fatal("authHeader() returned empty string - check .env credentials")
	}

	// For personal token, should be Basic auth with email:token
	if client.Token[:3] == "pt_" && len(header) > 6 {
		if header[:6] != "Basic " {
			t.Errorf("Expected 'Basic ' prefix for personal token auth, got: %s", header[:6])
		}
	} else if client.Token[:3] == "at_" && len(header) > 7 {
		if header[:7] != "Bearer " {
			t.Errorf("Expected 'Bearer ' prefix for OAuth token, got: %s", header[:7])
		}
	}

	fmt.Printf("Auth header is properly formatted: %s\n", header[:30])
}

func TestGetAvailability(t *testing.T) {
	loadEnvFile()
	client := NewClient()

	if client.BaseURL == "" || client.authHeader() == "" {
		t.Fatal("Credentials not configured")
	}

	// First get properties to find a property ID
	properties, err := client.GetProperties()
	if err != nil {
		t.Fatalf("GetProperties() failed: %v", err)
	}

	if len(properties) == 0 {
		t.Skip("No properties found, skipping availability test")
	}

	// Use the first property's ID (assuming it's a number)
	var propertyID int
	for _, prop := range properties {
		if idVal, ok := prop["id"].(int); ok {
			propertyID = idVal
			break
		}
		if idVal, ok := prop["id"].(float64); ok {
			propertyID = int(idVal)
			break
		}
	}

	fmt.Printf("Testing GetAvailability for property ID: %d\n", propertyID)

	// Test for July 2026
	availability, err := client.GetAvailability(propertyID, 2026, 7)
	if err != nil {
		t.Fatalf("GetAvailability() failed: %v", err)
	}

	fmt.Printf("Retrieved %d availability dates\n", len(availability))

	blockedCount := 0
	for _, date := range availability {
		if date.Blocked {
			blockedCount++
			fmt.Printf("  %s - %s\n", date.Date, date.Reason)
		}
	}

	fmt.Printf("\nTotal blocked dates: %d out of %d\n", blockedCount, len(availability))
}
