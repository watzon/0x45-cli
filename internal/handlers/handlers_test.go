package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/watzon/0x45-cli/internal/client"
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
		case "/urls":
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			resp := paste69.ListResponse[paste69.URLListItem]{
				Success: true,
			}
			resp.Data.Items = []paste69.URLListItem{
				{
					Id:          "abc123",
					URL:         "https://0x45.st/abc123",
					ShortURL:    "https://0x45.st/abc123",
					OriginalURL: "https://example.com",
					CreatedAt:   "2023-01-01T00:00:00Z",
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

func TestUploadHandler(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	viper.Set("api_url", server.URL)
	viper.Set("api_key", "test-key")
	client.Initialize()

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

	cmd := &cobra.Command{}
	cmd.Flags().Bool("private", true, "")
	cmd.Flags().String("expires", "24h", "")

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = Upload(cmd, []string{tmpfile.Name()})
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "https://0x45.st/abc123") {
		t.Error("Expected output to contain URL")
	}
}

func TestShortenHandler(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	viper.Set("api_url", server.URL)
	viper.Set("api_key", "test-key")
	client.Initialize()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("private", true, "")
	cmd.Flags().String("expires", "24h", "")

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := Shorten(cmd, []string{"https://example.com"})
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "https://0x45.st/abc123") {
		t.Error("Expected output to contain URL")
	}
}

func TestListPastesHandler(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	viper.Set("api_url", server.URL)
	viper.Set("api_key", "test-key")
	client.Initialize()

	cmd := &cobra.Command{}
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("per-page", 10, "")

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := List(cmd, []string{"pastes"})
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "test.txt") {
		t.Error("Expected output to contain filename")
	}
}

func TestListURLsHandler(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	viper.Set("api_url", server.URL)
	viper.Set("api_key", "test-key")
	client.Initialize()

	cmd := &cobra.Command{}
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("per-page", 10, "")

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := List(cmd, []string{"urls"})
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "https://example.com") {
		t.Error("Expected output to contain original URL")
	}
}

func TestDeleteHandler(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	viper.Set("api_url", server.URL)
	viper.Set("api_key", "test-key")
	client.Initialize()

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := Delete(cmd, []string{"abc123"})
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Deleted successfully") {
		t.Error("Expected output to contain success message")
	}
}
