package main

import (
	"fmt"

	"github.com/example/ownerrez-github-pages/internal/ownerrez"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	client := ownerrez.NewClient()

	// First get all properties to see what format they're in
	fmt.Println("--- All Properties (raw) ---")
	props, err := client.GetProperties()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, p := range props {
		fmt.Printf("Property: ID=%v, slug field=%v, public_url=%v\n", p["id"], p["slug"], p["public_url"])
	}

	// Test GetProperty by slug
	fmt.Println("\n--- Testing GetProperty ---")
	slugToTest := "sapala-all-exclusive-oceanfront-villa-with-private-pool-orp5b5eae5x"
	prop, err := client.GetProperty(slugToTest)
	if err != nil {
		fmt.Printf("Error getting property '%s': %v\n", slugToTest, err)
	} else {
		fmt.Printf("Found property: ID=%v, Slug=%v\n", prop["id"], prop["slug"])
	}

	// Test GetProperties again
	fmt.Println("\n--- Testing GetProperties (re-run) ---")
	props2, err := client.GetProperties()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Got %d properties\n", len(props2))
	for _, p := range props2 {
		fmt.Printf("ID: %v, Slug: %v, PublicURL: %v\n", p["id"], p["slug"], p["public_url"])
	}

	// Test GetPropertyDetails (if needed)
	fmt.Println("\n--- Testing GetPropertyDetails ---")
	details, err := client.GetPropertyDetails("sapala-all-exclusive-oceanfront-villa-with-private-pool-orp5b5eae5x")
	if err != nil {
		fmt.Printf("Error getting property details: %v\n", err)
	} else {
		fmt.Printf("Found details for: %s\n", details.Name)
		fmt.Printf("  Bedrooms: %d, Bathrooms: %d, Sleeps: %s\n", details.Bedrooms, details.Bathrooms, details.Sleeps)
		fmt.Printf("  Rate: $%.0f - $%.0f/night\n", details.NightlyRateMin, details.NightlyRateMax)
	}
}
