package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	baseURL := "https://example.com"
	apiKey := "test-key"
	client := New(baseURL, apiKey)

	if client.BaseUrl != baseURL {
		t.Errorf("Expected BaseUrl to be %s, got %s", baseURL, client.BaseUrl)
	}
	if client.APIKey != apiKey {
		t.Errorf("Expected APIKey to be %s, got %s", apiKey, client.APIKey)
	}
}

func TestUpload(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify auth header
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("Expected Authorization header 'Bearer test-key', got %s", auth)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		resp := UploadResponse{
			Success: true,
			Data: struct {
				Id          string     `json:"id"`
				Url         string     `json:"url"`
				RawUrl      string     `json:"raw_url"`
				DownloadUrl string     `json:"download_url"`
				DeleteUrl   string     `json:"delete_url"`
				Filename    string     `json:"filename"`
				MimeType    string     `json:"mime_type"`
				Size        int64      `json:"size"`
				Private     bool       `json:"private"`
				CreatedAt   time.Time  `json:"created_at"`
				ExpiresAt   *time.Time `json:"expires_at"`
			}{
				Id:          "test123",
				Url:         "https://0x45.st/test123",
				RawUrl:      "https://0x45.st/raw/test123",
				DownloadUrl: "https://0x45.st/download/test123",
				DeleteUrl:   "https://0x45.st/delete/test123",
				Filename:    "test.txt",
				MimeType:    "text/plain",
				Size:        12,
				Private:     false,
				CreatedAt:   time.Now(),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	content := bytes.NewBufferString("test content")
	query := url.Values{}
	query.Set("filename", "test.txt")

	resp, err := client.Upload(content, query)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Data.Id != "test123" {
		t.Errorf("Expected Id to be test123, got %s", resp.Data.Id)
	}
}

func TestShorten(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		resp := ShortenResponse{
			Success: true,
			Data: struct {
				Id        string     `json:"id"`
				ShortUrl  string     `json:"short_url"`
				Url       string     `json:"url"`
				Title     string     `json:"title"`
				DeleteUrl string     `json:"delete_url"`
				Clicks    int        `json:"clicks"`
				LastClick *time.Time `json:"last_click"`
				CreatedAt time.Time  `json:"created_at"`
				ExpiresAt *time.Time `json:"expires_at"`
			}{
				Id:        "abc123",
				ShortUrl:  "https://0x45.st/abc123",
				Url:       "https://example.com",
				Title:     "Test URL",
				DeleteUrl: "https://0x45.st/delete/abc123",
				Clicks:    0,
				CreatedAt: time.Now(),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	opts := ShortenOptions{
		Url:   "https://example.com",
		Title: "Test URL",
	}

	resp, err := client.Shorten(opts)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Data.Id != "abc123" {
		t.Errorf("Expected Id to be abc123, got %s", resp.Data.Id)
	}
}

func TestListPastes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
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
						Id:       "test123",
						Filename: "test.txt",
						Size:     12,
						MimeType: "text/plain",
					},
				},
				Total: 1,
				Page:  1,
				Limit: 10,
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	opts := ListOptions{
		Page:  1,
		Limit: 10,
	}

	resp, err := client.ListPastes(opts)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if len(resp.Data.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(resp.Data.Items))
	}
}

func TestDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		if r.URL.Path != "/test123" {
			t.Errorf("Expected path /test123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		resp := DeleteResponse{
			Success: true,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	resp, err := client.Delete("test123")
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Success {
		t.Error("Expected success response")
	}
}
