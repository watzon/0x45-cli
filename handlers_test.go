package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Helper functions for testing
func setupTestConfig(t *testing.T) func() {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".0x45.yaml")
	viper.SetConfigFile(configFile)
	viper.Set("api_url", "http://test-api")
	viper.Set("api_key", "test-key")
	err := viper.WriteConfig()
	if err != nil {
		t.Fatal(err)
	}

	return func() {
		viper.Reset()
		os.RemoveAll(tmpDir)
	}
}

func TestHandleConfigSet(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Create a new command with output capture
	cmd := &cobra.Command{Use: "config"}
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	handleConfigSet(cmd, []string{"test_key", "test_value"})

	if !viper.IsSet("test_key") {
		t.Error("Expected test_key to be set")
	}
	if value := viper.GetString("test_key"); value != "test_value" {
		t.Errorf("Expected test_key to be 'test_value', got '%s'", value)
	}
	if output.Len() == 0 {
		t.Error("Expected output to contain success message")
	}
}

func TestHandleConfigGet(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Create a new command with output capture
	cmd := &cobra.Command{Use: "config"}
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	
	// Test existing key
	handleConfigGet(cmd, []string{"api_key"})
	if !bytes.Contains(output.Bytes(), []byte("test-key")) {
		t.Errorf("Expected output to contain 'test-key', got '%s'", output.String())
	}

	// Test non-existing key
	output.Reset()
	handleConfigGet(cmd, []string{"nonexistent_key"})
	if !bytes.Contains(output.Bytes(), []byte("Config key 'nonexistent_key' not found")) {
		t.Errorf("Expected not found message, got '%s'", output.String())
	}
}

func TestHandleUpload(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("Expected Authorization header 'Bearer test-key', got %s", auth)
		}

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
				Id:       "test123",
				Url:      "https://0x45.st/test123",
				Filename: "test.txt",
				MimeType: "text/plain",
				Size:     12,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	// Set API URL to test server
	viper.Set("api_url", server.URL)

	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("test content"); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cmd := &cobra.Command{Use: "upload"}
	cmd.Flags().Bool("private", false, "")
	cmd.Flags().String("expires", "", "")

	err = handleUpload(cmd, []string{tmpFile.Name()})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHandleShorten(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

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
				Id:       "abc123",
				ShortUrl: "https://0x45.st/abc123",
				Url:      "https://example.com",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	viper.Set("api_url", server.URL)

	cmd := &cobra.Command{Use: "shorten"}
	cmd.Flags().String("expires", "", "")
	cmd.Flags().String("title", "", "")

	err := handleShorten(cmd, []string{"https://example.com"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHandleDelete(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
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

	viper.Set("api_url", server.URL)

	cmd := &cobra.Command{Use: "delete"}
	err := handleDelete(cmd, []string{"test123"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHandleListUrls(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		resp := ListResponse[UrlListItem]{
			Success: true,
			Data: struct {
				Items []UrlListItem `json:"items"`
				Total int           `json:"total"`
				Page  int           `json:"page"`
				Limit int           `json:"limit"`
			}{
				Items: []UrlListItem{
					{
						Id:       "abc123",
						ShortUrl: "https://0x45.st/abc123",
						Url:      "https://example.com",
						Clicks:   5,
					},
				},
				Total: 1,
				Page:  1,
				Limit: 10,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	viper.Set("api_url", server.URL)

	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().Int("limit", 10, "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().String("sort", "", "")

	err := handleListUrls(cmd, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHandleListPastes(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		resp := ListResponse[PasteListItem]{
			Success: true,
			Data: struct {
				Items []PasteListItem `json:"items"`
				Total int            `json:"total"`
				Page  int            `json:"page"`
				Limit int            `json:"limit"`
			}{
				Items: []PasteListItem{
					{
						Id:          "test123",
						Filename:    "test.txt",
						Size:        12,
						MimeType:    "text/plain",
						Url:         "https://0x45.st/test123",
						RawUrl:      "https://0x45.st/raw/test123",
						DownloadUrl: "https://0x45.st/download/test123",
						DeleteUrl:   "https://0x45.st/delete/test123",
						CreatedAt:   time.Now(),
					},
				},
				Total: 1,
				Page:  1,
				Limit: 10,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	viper.Set("api_url", server.URL)

	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().Int("limit", 10, "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().String("sort", "", "")

	err := handleListPastes(cmd, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
