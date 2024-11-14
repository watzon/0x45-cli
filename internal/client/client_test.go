package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/watzon/0x45-cli/pkg/api/paste69"
)

func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/upload":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			resp := paste69.UploadResponse{
				Success:   true,
				URL:       "https://0x45.st/abc123",
				DeleteURL: "https://0x45.st/delete/abc123",
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "/shorten":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			resp := paste69.ShortenResponse{
				Success:   true,
				URL:       "https://0x45.st/abc123",
				DeleteURL: "https://0x45.st/delete/abc123",
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "/pastes":
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			resp := paste69.ListResponse[paste69.PasteListItem]{
				Success: true,
			}
			resp.Data.Items = []paste69.PasteListItem{
				{
					Id:        "abc123",
					Filename:  "test.txt",
					Size:      123,
					CreatedAt: "2023-01-01T00:00:00Z",
					URL:       "https://0x45.st/abc123",
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "/delete/abc123":
			if r.Method != http.MethodDelete {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			resp := paste69.GenericResponse{
				Success: true,
				Message: "Deleted successfully",
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestUploadFile(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Initialize a new client for each test
	client = paste69.NewClient(server.URL, "test-key")

	// Create a temporary test file
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("test content")); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	resp, err := UploadFile(tmpfile.Name(), true, "24h")
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
	if resp.URL != "https://0x45.st/abc123" {
		t.Errorf("Expected URL to be https://0x45.st/abc123, got %s", resp.URL)
	}
}

func TestShortenURL(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Initialize a new client for each test
	client = paste69.NewClient(server.URL, "test-key")

	resp, err := ShortenURL("https://example.com", true, "24h")
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
	if resp.URL != "https://0x45.st/abc123" {
		t.Errorf("Expected URL to be https://0x45.st/abc123, got %s", resp.URL)
	}
}

func TestListPastes(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Initialize a new client for each test
	client = paste69.NewClient(server.URL, "test-key")

	resp, err := ListPastes(1, 10)
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
	if len(resp.Data.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(resp.Data.Items))
	}
}

func TestDelete(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Initialize a new client for each test
	client = paste69.NewClient(server.URL, "test-key")

	resp, err := Delete("abc123")
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
	if resp.Message != "Deleted successfully" {
		t.Errorf("Expected message to be 'Deleted successfully', got %s", resp.Message)
	}
}
