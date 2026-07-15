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
