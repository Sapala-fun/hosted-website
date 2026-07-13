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
}

func NewClient() *Client {
    return &Client{
        BaseURL: os.Getenv("OWNERREZ_API_BASE_URL"),
        APIKey:  os.Getenv("OWNERREZ_API_KEY"),
    }
}

func (c *Client) GetProperties() ([]map[string]any, error) {
    if c.BaseURL == "" || c.APIKey == "" {
        return nil, fmt.Errorf("ownerrez credentials not configured")
    }

    req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/v1/properties", nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-API-Key", c.APIKey)
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
        Properties []map[string]any `json:"properties"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
        return nil, err
    }

    return payload.Properties, nil
}

func (c *Client) CreateBooking(payload map[string]string) (map[string]any, error) {
    if c.BaseURL == "" || c.APIKey == "" {
        return nil, fmt.Errorf("ownerrez credentials not configured")
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v1/bookings", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-API-Key", c.APIKey)
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
