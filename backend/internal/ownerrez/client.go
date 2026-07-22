package ownerrez

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	BaseURL    string
	APIKey     string
	Token      string // Personal token (current), will switch to OAuth later
	OwnerEmail string
}

// stripQuotes removes surrounding double quotes from a string if present
func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func NewClient() *Client {
	return &Client{
		BaseURL:    stripQuotes(os.Getenv("OWNERREZ_API_BASE_URL")),
		APIKey:     stripQuotes(os.Getenv("OWNERREZ_API_KEY")),
		Token:      stripQuotes(os.Getenv("OWNERREZ_PERSONAL_TOKEN")),
		OwnerEmail: stripQuotes(os.Getenv("OWNERREZ_EMAIL")),
	}
}

// authHeader returns the appropriate Authorization header based on available credentials.
// Priority for Personal Access Tokens (pt_...): Basic Auth (email + token) > OAuth Token
// Priority for API Keys: HTTP Basic (base64 encoded)
func (c *Client) authHeader() string {
	// If we have a personal access token, use HTTP Basic Auth with email as username
	if c.Token != "" && c.Token[:3] == "pt_" && c.OwnerEmail != "" {
		auth := c.OwnerEmail + ":" + c.Token
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	}
	if c.Token != "" && c.Token[:3] == "at_" {
		// OAuth Token (bearer)
		return "Bearer " + c.Token
	}
	if c.APIKey != "" {
		// API Key for HTTP Basic auth - already base64 encoded by user
		return "Basic " + c.APIKey
	}
	return ""
}

func (c *Client) GetProperties() ([]map[string]any, error) {
	if c.BaseURL == "" || c.authHeader() == "" {
		return nil, fmt.Errorf("ownerrez credentials not configured")
	}

	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/v2/properties", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ownerrez properties request failed: %s", resp.Status)
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	// Normalize property data - convert image_url to imageUrl for frontend compatibility
	for i := range payload.Items {
		prop := &payload.Items[i]

		// Handle various possible image field names from OwnerRez API
		var imageUrl string
		if img, ok := (*prop)["image_url"].(string); ok && img != "" {
			imageUrl = img
		} else if img, ok := (*prop)["imageUrl"].(string); ok && img != "" {
			imageUrl = img
		} else if img, ok := (*prop)["hero_image"].(string); ok && img != "" {
			imageUrl = img
		}

		if imageUrl != "" {
			(*prop)["imageUrl"] = imageUrl
		}

		// Also normalize nightlyRate fields
		if rate, ok := (*prop)["nightly_rate"].(float64); ok {
			(*prop)["nightlyRate"] = rate
		}
		if rateMin, ok := (*prop)["nightly_rate_min"].(float64); ok {
			(*prop)["nightlyRateMin"] = rateMin
		}
		if rateMax, ok := (*prop)["nightly_rate_max"].(float64); ok {
			(*prop)["nightlyRateMax"] = rateMax
		}

		// Convert id from float64 to int-like format for consistency
		if id, ok := (*prop)["id"].(float64); ok {
			(*prop)["id"] = int(id)
		}
	}

	return payload.Items, nil
}

func (c *Client) GetProperty(slug string) (map[string]any, error) {
	if c.BaseURL == "" || c.authHeader() == "" {
		return nil, fmt.Errorf("ownerrez credentials not configured")
	}

	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/v2/properties", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ownerrez properties request failed: %s", resp.Status)
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	for _, prop := range payload.Items {
		// Check for slug field first (if present in API response)
		if s, ok := prop["slug"].(string); ok && s == slug {
			return prop, nil
		}
		// Fallback: check public_url for slug match (OwnerRez API doesn't return explicit slug)
		expectedUrl := fmt.Sprintf("https://www.sapala.fun/%s", slug)
		if s, ok := prop["public_url"].(string); ok {
			if len(s) > len(expectedUrl) && s[:len(expectedUrl)] == expectedUrl {
				return prop, nil
			}
		}
	}
	return nil, fmt.Errorf("property not found: %s", slug)
}

func (c *Client) CreateBooking(payload map[string]string) (map[string]any, error) {
	if c.BaseURL == "" || c.authHeader() == "" {
		return nil, fmt.Errorf("ownerrez credentials not configured")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v2/bookings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ownerrez booking request failed: %s", resp.Status)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// AvailabilityDate represents a single date in the availability calendar.
type AvailabilityDate struct {
	Date        string  `json:"date"`
	Blocked     bool    `json:"blocked"`
	Reason      string  `json:"reason,omitempty"`
	NightlyRate float64 `json:"nightlyRate,omitempty"`
}

// PropertyDetails represents enriched property data fetched from OwnerRez.
type PropertyDetails struct {
	ID                 string              `json:"id"`
	Slug               string              `json:"slug"`
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	Bedrooms           int                 `json:"bedrooms"`
	Bathrooms          int                 `json:"bathrooms"`
	Sleeps             string              `json:"sleeps"`
	NightlyRateMin     float64             `json:"nightlyRateMin"`
	NightlyRateMax     float64             `json:"nightlyRateMax"`
	ImageUrl           string              `json:"imageUrl"`
	Photos             []string            `json:"photos"`
	Amenities          []map[string]string `json:"amenities"`
	CancellationPolicy string              `json:"cancellationPolicy"`
	OwnerName          string              `json:"ownerName"`
	OwnerBio           string              `json:"ownerBio"`
	Location           string              `json:"location"`
	BookingURL         string              `json:"bookingUrl"`
	GuestReviews       []map[string]string `json:"guestReviews"`
}

// GetPropertyDetails fetches enriched property details from OwnerRez API.
func (c *Client) GetPropertyDetails(slug string) (*PropertyDetails, error) {
	if c.BaseURL == "" || c.authHeader() == "" {
		return nil, fmt.Errorf("ownerrez credentials not configured")
	}

	// First, get the list of properties to find the property ID
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/v2/properties", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	respList, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer respList.Body.Close()

	if respList.StatusCode >= 400 {
		return nil, fmt.Errorf("ownerrez properties request failed: %s", respList.Status)
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(respList.Body).Decode(&payload); err != nil {
		return nil, err
	}

	expectedUrl := fmt.Sprintf("https://www.sapala.fun/%s", slug)

	// Find the property by matching the public_url
	var propID int
	var propSlug string
	for _, prop := range payload.Items {
		if s, ok := prop["slug"].(string); ok && s == slug {
			propSlug = s
			propID, _ = prop["id"].(int)
			break
		}
		if pubUrl, ok := prop["public_url"].(string); ok {
			// Direct exact match (slug might already include the -or suffix)
			if pubUrl == expectedUrl {
				propSlug = slug
				// Try to get ID as float64 first (JSON numbers often decode as float64)
				idVal := prop["id"]
				switch v := idVal.(type) {
				case int:
					propID = v
				case float64:
					propID = int(v)
				case string:
					// Try to parse as int
					fmt.Sscanf(v, "%d", &propID)
				}
				break
			}
			// Prefix match: expectedUrl is a prefix of pubUrl
			if len(pubUrl) > len(expectedUrl) && pubUrl[:len(expectedUrl)] == expectedUrl {
				propSlug = slug
				propID, _ = prop["id"].(int)
				break
			}
			// Fallback: match up to the -or in public_url
			if idx := strings.Index(pubUrl, "-or"); idx > 3 && idx+len("-or") < len(pubUrl) {
				if prefix := pubUrl[:idx]; prefix == expectedUrl {
					propSlug = slug
					propID, _ = prop["id"].(int)
					break
				}
			} else if idx := strings.Index(pubUrl, "-or"); idx > 3 && idx+len("-or") == len(pubUrl) {
				if prefix := pubUrl[:idx]; prefix == expectedUrl {
					propSlug = slug
					propID, _ = prop["id"].(int)
					break
				}
			}
		}
	}

	if propID == 0 || propSlug == "" {
		return nil, fmt.Errorf("property not found: %s (matched propID=%d, propSlug='%s')", slug, propID, propSlug)
	}

	// Now fetch the full property details using /v2/properties/{id}
	detailReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/properties/%d", c.BaseURL, propID), nil)
	if err != nil {
		return nil, err
	}
	detailReq.Header.Set("Authorization", c.authHeader())
	detailReq.Header.Set("Content-Type", "application/json")

	respDetail, err := http.DefaultClient.Do(detailReq)
	if err != nil {
		return nil, err
	}
	defer respDetail.Body.Close()

	if respDetail.StatusCode >= 400 {
		return nil, fmt.Errorf("ownerrez property details request failed: %s", respDetail.Status)
	}

	var prop map[string]any
	if err := json.NewDecoder(respDetail.Body).Decode(&prop); err != nil {
		return nil, err
	}

	// Extract photos from images array if available
	var photos []string
	if imgs, ok := prop["images"].([]any); ok {
		for _, img := range imgs {
			if imgStr, ok := img.(string); ok {
				photos = append(photos, imgStr)
			}
		}
	} else if img, ok := prop["image_url"].(string); ok && img != "" {
		photos = append(photos, img)
	}

	// Extract amenities from property data
	var amenities []map[string]string
	if amens, ok := prop["amenities"].([]any); ok {
		for _, a := range amens {
			if amMap, ok := a.(map[string]any); ok {
				amStr := make(map[string]string)
				for k, v := range amMap {
					amStr[k] = fmt.Sprintf("%v", v)
				}
				amenities = append(amenities, amStr)
			}
		}
	}

	// Get nightly rates
	nightlyRateMin := 0.0
	nightlyRateMax := 0.0
	if nr, ok := prop["nightly_rate_min"].(float64); ok {
		nightlyRateMin = nr
	} else if nr, ok := prop["nightlyRate"].(float64); ok {
		nightlyRateMin = nr
	}
	if nrMax, ok := prop["nightly_rate_max"].(float64); ok {
		nightlyRateMax = nrMax
	}

	// Get sleeps as string
	sleeps := ""
	if s, ok := prop["sleeps"].(string); ok {
		sleeps = s
	} else if s, ok := prop["sleeps"].(float64); ok {
		sleeps = fmt.Sprintf("%d", int(s))
	}

	// Get bedrooms/bathrooms as ints
	bedrooms := 0
	if b, ok := prop["bedrooms"].(float64); ok {
		bedrooms = int(b)
	}
	bathrooms := 0
	if b, ok := prop["bathrooms"].(float64); ok {
		bathrooms = int(b)
	}

	// Get cancellation policy if available
	cancellationPolicy := ""
	if cp, ok := prop["cancellation_policy"].(string); ok {
		cancellationPolicy = cp
	} else if cpMap, ok := prop["cancellationPolicy"].(map[string]any); ok {
		if desc, ok := cpMap["description"].(string); ok {
			cancellationPolicy = desc
		}
	}

	// Get owner info if available
	ownerName := ""
	if owner, ok := prop["owner"].(map[string]any); ok {
		if name, ok := owner["name"].(string); ok {
			ownerName = name
		}
	} else if name, ok := prop["owner_name"].(string); ok {
		ownerName = name
	}
	ownerBio := ""
	if bio, ok := prop["owner_bio"].(string); ok {
		ownerBio = bio
	} else if owner, ok := prop["owner"].(map[string]any); ok {
		if bio, ok := owner["bio"].(string); ok {
			ownerBio = bio
		}
	}

	// Get location if available
	location := "Christiansted, St. Croix, USVI, Virgin Islands (U.S.)"
	if loc, ok := prop["location"].(string); ok {
		location = loc
	} else if addr, ok := prop["address"].(map[string]any); ok {
		var parts []string
		if city, ok := addr["city"].(string); ok {
			parts = append(parts, city)
		}
		if region, ok := addr["region"].(string); ok {
			parts = append(parts, region)
		}
		if country, ok := addr["country"].(string); ok {
			parts = append(parts, country)
		}
		if len(parts) > 0 {
			location = parts[0]
			for i := 1; i < len(parts); i++ {
				location += ", " + parts[i]
			}
		}
	}

	// Get booking URL if available
	bookingURL := ""
	if bu, ok := prop["booking_url"].(string); ok {
		bookingURL = bu
	} else if bu, ok := prop["bookOnlineUrl"].(string); ok {
		bookingURL = bu
	}

	// Get guest reviews if available
	var guestReviews []map[string]string
	if reviews, ok := prop["reviews"].([]any); ok {
		for _, r := range reviews {
			if revMap, ok := r.(map[string]any); ok {
				revStr := make(map[string]string)
				for k, v := range revMap {
					revStr[k] = fmt.Sprintf("%v", v)
				}
				guestReviews = append(guestReviews, revStr)
			}
		}
	}

	// Get image URL for hero
	imageUrl := ""
	if img, ok := prop["image_url"].(string); ok {
		imageUrl = img
	} else if len(photos) > 0 {
		imageUrl = photos[0]
	}

	// Try multiple possible description field names
	description := ""
	descFields := []string{"description", "description_html", "long_description", "Overview"}
	for _, field := range descFields {
		if d, ok := prop[field].(string); ok && d != "" {
			description = d
			break
		}
	}

	return &PropertyDetails{
		ID:                 fmt.Sprintf("%d", propID),
		Slug:               propSlug,
		Name:               getString(prop, "name"),
		Description:        description,
		Bedrooms:           bedrooms,
		Bathrooms:          bathrooms,
		Sleeps:             sleeps,
		NightlyRateMin:     nightlyRateMin,
		NightlyRateMax:     nightlyRateMax,
		ImageUrl:           imageUrl,
		Photos:             photos,
		Amenities:          amenities,
		CancellationPolicy: cancellationPolicy,
		OwnerName:          ownerName,
		OwnerBio:           ownerBio,
		Location:           location,
		BookingURL:         bookingURL,
		GuestReviews:       guestReviews,
	}, nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// GetAvailability fetches bookings for a property and derives blocked/available dates
// for the given month (year, month). Blocked dates come from active/pending bookings
// or blocks; available dates include the nightly rate.
func (c *Client) GetAvailability(propertyID int, year int, month int) ([]AvailabilityDate, error) {
	if c.BaseURL == "" || c.authHeader() == "" {
		return nil, fmt.Errorf("ownerrez credentials not configured")
	}

	// Build date range for the month
	start := fmt.Sprintf("%04d-%02d-01T00:00:00", year, month)
	// Get last day of month
	lastDay := 31
	switch month {
	case 4, 6, 9, 11:
		lastDay = 30
	case 2:
		lastDay = 28
	}
	end := fmt.Sprintf("%04d-%02d-%02dT23:59:59", year, month, lastDay)

	url := fmt.Sprintf("%s/v2/bookings?property_ids=%d&from=%s&to=%s&include_guest=true&limit=100",
		c.BaseURL, propertyID, start, end)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ownerrez availability request failed: %s", resp.Status)
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	// Build a map of blocked dates from bookings
	blockedDates := make(map[string]string) // date -> reason
	for _, item := range payload.Items {
		status, _ := item["status"].(string)
		if status != "active" && status != "pending" {
			continue
		}

		arrivalRaw, hasArrival := item["arrival"]
		departureRaw, hasDeparture := item["departure"]
		isBlock, _ := item["is_block"].(bool)
		title, _ := item["title"].(string)

		if !hasArrival || !hasDeparture {
			continue
		}

		arrivalStr, _ := arrivalRaw.(string)
		departureStr, _ := departureRaw.(string)

		reason := title
		if isBlock {
			reason = "Blocked"
		} else if reason == "" {
			reason = "Reserved"
		}

		// Parse arrival and departure dates
		var arrivalDate, departureDate string
		if len(arrivalStr) >= 10 {
			arrivalDate = arrivalStr[:10]
		}
		if len(departureStr) >= 10 {
			departureDate = departureStr[:10]
		}

		if arrivalDate != "" && departureDate != "" {
			for d := arrivalDate; d <= departureDate; d = nextDay(d) {
				if _, exists := blockedDates[d]; !exists {
					blockedDates[d] = reason
				}
			}
		}
	}

	// Build the calendar with available + blocked dates
	var dates []AvailabilityDate
	for day := 1; day <= lastDay; day++ {
		dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		if reason, ok := blockedDates[dateStr]; ok {
			dates = append(dates, AvailabilityDate{
				Date:    dateStr,
				Blocked: true,
				Reason:  reason,
			})
		} else {
			dates = append(dates, AvailabilityDate{
				Date:    dateStr,
				Blocked: false,
			})
		}
	}

	return dates, nil
}

// nextDay returns the next day in YYYY-MM-DD format.
func nextDay(dateStr string) string {
	var y, m, d int
	fmt.Sscanf(dateStr, "%04d-%02d-%02d", &y, &m, &d)
	d++
	switch m {
	case 4, 6, 9, 11:
		if d > 30 {
			d = 1
			m++
		}
	case 2:
		if d > 28 {
			d = 1
			m++
		}
	default:
		if d > 31 {
			d = 1
			m++
		}
	}
	if m > 12 {
		m = 1
		y++
	}
	return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
}
