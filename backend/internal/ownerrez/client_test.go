package ownerrez

import (
	"testing"
)

func TestAuthHeader_PersonalToken(t *testing.T) {
	client := &Client{
		BaseURL: "https://api.ownerrez.com",
		Token:   "test-token-123",
	}

	header := client.authHeader()
	expected := "Bearer test-token-123"
	if header != expected {
		t.Errorf("authHeader() = %q, want %q", header, expected)
	}
}

func TestAuthHeader_APIKey(t *testing.T) {
	client := &Client{
		BaseURL: "https://api.ownerrez.com",
		APIKey:  "test-api-key-456",
	}

	header := client.authHeader()
	expected := "Basic test-api-key-456"
	if header != expected {
		t.Errorf("authHeader() = %q, want %q", header, expected)
	}
}

func TestAuthHeader_PersonalTokenPriority(t *testing.T) {
	client := &Client{
		BaseURL: "https://api.ownerrez.com",
		Token:   "test-token-123",
		APIKey:  "test-api-key-456",
	}

	header := client.authHeader()
	expected := "Bearer test-token-123"
	if header != expected {
		t.Errorf("authHeader() = %q, want %q (personal token should have priority)", header, expected)
	}
}

func TestAuthHeader_Empty(t *testing.T) {
	client := &Client{
		BaseURL: "https://api.ownerrez.com",
	}

	header := client.authHeader()
	if header != "" {
		t.Errorf("authHeader() = %q, want empty string", header)
	}
}

func TestNewClient_EnvVars(t *testing.T) {
	t.Setenv("OWNERREZ_API_BASE_URL", "https://api.ownerrez.com")
	t.Setenv("OWNERREZ_API_KEY", "test-key")
	t.Setenv("OWNERREZ_PERSONAL_TOKEN", "test-token")

	client := NewClient()

	if client.BaseURL != "https://api.ownerrez.com" {
		t.Errorf("BaseURL = %q, want %q", client.BaseURL, "https://api.ownerrez.com")
	}
	if client.APIKey != "test-key" {
		t.Errorf("APIKey = %q, want %q", client.APIKey, "test-key")
	}
	if client.Token != "test-token" {
		t.Errorf("Token = %q, want %q", client.Token, "test-token")
	}
}

func TestNewClient_NoEnvVars(t *testing.T) {
	client := NewClient()

	if client.BaseURL != "" {
		t.Errorf("BaseURL = %q, want empty", client.BaseURL)
	}
	if client.APIKey != "" {
		t.Errorf("APIKey = %q, want empty", client.APIKey)
	}
	if client.Token != "" {
		t.Errorf("Token = %q, want empty", client.Token)
	}
}

func TestGetProperties_NoCredentials(t *testing.T) {
	client := &Client{}
	_, err := client.GetProperties()

	if err == nil {
		t.Error("GetProperties() expected error with no credentials, got nil")
	}
}

func TestCreateBooking_NoCredentials(t *testing.T) {
	client := &Client{}
	_, err := client.CreateBooking(map[string]string{
		"guestName": "Test User",
		"checkIn":   "2026-08-01",
		"checkOut":  "2026-08-03",
	})

	if err == nil {
		t.Error("CreateBooking() expected error with no credentials, got nil")
	}
}
