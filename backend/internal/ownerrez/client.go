package ownerrez

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Client struct {
	BaseURL string
	APIKey  string
	Token   string // Personal token (current), will switch to OAuth later
}

func NewClient() *Client {
	return &Client{
		BaseURL: os.Getenv("OWNERREZ_API_BASE_URL"),
		APIKey:  os.Getenv("OWNERREZ_API_KEY"),
		Token:   os.Getenv("OWNERREZ_PERSONAL_TOKEN"),
	}
}

// authHeader returns the appropriate Authorization header based on available credentials.
// Priority: Personal Token > API Key (HTTP Basic) > OAuth Token
func (c *Client) authHeader() string {
	if c.Token != "" {
		return "Bearer " + c.Token
	}
	if c.APIKey != "" {
		return "Basic " + c.APIKey
	}
	if c.Token != "" {
		return "Bearer " + c.Token
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
		Data []map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload.Data, nil
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
		Data []map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	for _, prop := range payload.Data {
		if s, ok := prop["slug"].(string); ok && s == slug {
			return prop, nil
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
	Date     string `json:"date"`
	Blocked  bool   `json:"blocked"`
	Reason   string `json:"reason,omitempty"`
	NightlyRate float64 `json:"nightlyRate,omitempty"`
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
