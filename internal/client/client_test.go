package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestUploadFile(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/upload" {
			t.Errorf("Expected /upload path, got %s", r.URL.Path)
		}

		// Check query parameters
		if r.URL.Query().Get("private") != "true" {
			t.Error("Expected private=true in query")
		}
		if r.URL.Query().Get("expires") != "24h" {
			t.Error("Expected expires=24h in query")
		}

		// Return mock response
		resp := UploadResponse{
			Success:   true,
			URL:       "https://0x45.st/abc123",
			DeleteURL: "https://0x45.st/delete/abc123",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Set up viper config for test
	viper.Set("api_url", server.URL)
	viper.Set("api_key", "test-key")

	// Create a temporary test file
	content := []byte("test content")
	tmpfile := t.TempDir() + "/test.txt"
	if err := os.WriteFile(tmpfile, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Test upload
	resp, err := UploadFile(tmpfile, true, "24h")
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Success {
		t.Error("Expected successful response")
	}
	if resp.URL != "https://0x45.st/abc123" {
		t.Errorf("Expected URL https://0x45.st/abc123, got %s", resp.URL)
	}
}

func TestShortenURL(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/shorten" {
			t.Errorf("Expected /shorten path, got %s", r.URL.Path)
		}

		// Check query parameters
		if r.URL.Query().Get("url") != "https://example.com" {
			t.Error("Expected url=https://example.com in query")
		}

		// Return mock response
		resp := ShortenResponse{
			Success:   true,
			URL:       "https://0x45.st/abc123",
			DeleteURL: "https://0x45.st/delete/abc123",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Set up viper config for test
	viper.Set("api_url", server.URL)
	viper.Set("api_key", "test-key")

	// Test shorten
	resp, err := ShortenURL("https://example.com", false, "")
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Success {
		t.Error("Expected successful response")
	}
	if resp.URL != "https://0x45.st/abc123" {
		t.Errorf("Expected URL https://0x45.st/abc123, got %s", resp.URL)
	}
}

func TestDelete(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/delete/abc123" {
			t.Errorf("Expected /delete/abc123 path, got %s", r.URL.Path)
		}

		// Return mock response
		resp := DeleteResponse{
			Success: true,
			Message: "Content deleted successfully",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Set up viper config for test
	viper.Set("api_url", server.URL)
	viper.Set("api_key", "test-key")

	// Test delete
	resp, err := Delete("abc123")
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Success {
		t.Error("Expected successful response")
	}
	if resp.Message != "Content deleted successfully" {
		t.Errorf("Expected message 'Content deleted successfully', got %s", resp.Message)
	}
}

func TestListPastes(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/pastes" {
			t.Errorf("Expected /pastes path, got %s", r.URL.Path)
		}

		// Return mock response
		resp := ListResponse[PasteListItem]{
			Success: true,
			Data: struct {
				Items []PasteListItem `json:"items"`
				Total int             `json:"total"`
				Page  int             `json:"page"`
				Limit int             `json:"limit"`
			}{
				Items: []PasteListItem{
					{
						Id:        "abc123",
						Filename:  "test.txt",
						Size:      100,
						CreatedAt: "2024-01-01",
						URL:       "https://0x45.st/abc123",
					},
				},
				Total: 1,
				Page:  1,
				Limit: 10,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Set up viper config for test
	viper.Set("api_url", server.URL)
	viper.Set("api_key", "test-key")

	// Test list pastes
	resp, err := ListPastes(1, 10)
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Success {
		t.Error("Expected successful response")
	}
	if len(resp.Data.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(resp.Data.Items))
	}
	if resp.Data.Items[0].Id != "abc123" {
		t.Errorf("Expected ID abc123, got %s", resp.Data.Items[0].Id)
	}
}
