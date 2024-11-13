package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
}

type UploadOptions struct {
	Filename string
	Expires  string
	Private  bool
}

type UploadResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID          string     `json:"id"`
		URL         string     `json:"url"`
		RawURL      string     `json:"raw_url"`
		DownloadURL string     `json:"download_url"`
		DeleteURL   string     `json:"delete_url"`
		Filename    string     `json:"filename"`
		MimeType    string     `json:"mime_type"`
		Size        int64      `json:"size"`
		Private     bool       `json:"private"`
		CreatedAt   time.Time  `json:"created_at"`
		ExpiresAt   *time.Time `json:"expires_at"`
	} `json:"data"`
}

type ShortenOptions struct {
	URL     string
	Expires string
	Title   string
}

type ShortenResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID        string     `json:"id"`
		ShortURL  string     `json:"short_url"`
		URL       string     `json:"url"`
		Title     string     `json:"title"`
		DeleteURL string     `json:"delete_url"`
		Clicks    int        `json:"clicks"`
		LastClick *time.Time `json:"last_click"`
		CreatedAt time.Time  `json:"created_at"`
		ExpiresAt *time.Time `json:"expires_at"`
	} `json:"data"`
}

type ListOptions struct {
	Type  string
	Page  int
	Limit int
	Sort  string
}

type ListItem struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	ShortURL  string    `json:"short_url"`
	Type      string    `json:"type"`  // "paste" or "url"
	Title     string    `json:"title"` // For URLs, stores original URL
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Size      int64     `json:"size"`   // For pastes
	Clicks    int       `json:"clicks"` // For URLs
	DeleteID  string    `json:"delete_id"`
}

type ListResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Items []ListItem `json:"items"`
		Total int        `json:"total"`
		Page  int        `json:"page"`
		Limit int        `json:"limit"`
	} `json:"data"`
}

type KeyRequestOptions struct {
	Email string
	Name  string
}

type KeyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type URLStatsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID        string     `json:"id"`
		URL       string     `json:"url"`
		ShortURL  string     `json:"short_url"`
		Clicks    int        `json:"clicks"`
		LastClick *time.Time `json:"last_click"`
		CreatedAt time.Time  `json:"created_at"`
		ExpiresAt *time.Time `json:"expires_at"`
	} `json:"data"`
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
}

func (c *Client) Upload(content io.Reader, opts UploadOptions) (*UploadResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file
	part, err := writer.CreateFormFile("file", filepath.Base(opts.Filename))
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}

	// Copy content to form file
	if _, err := io.Copy(part, content); err != nil {
		return nil, fmt.Errorf("copying content: %w", err)
	}

	// Add other form fields
	if opts.Expires != "" {
		writer.WriteField("expires", opts.Expires)
	}
	if opts.Private {
		writer.WriteField("private", "true")
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.BaseURL, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed: %s: %s", resp.Status, string(body))
	}

	// Parse JSON response
	var result UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}

func (c *Client) Shorten(opts ShortenOptions) (*ShortenResponse, error) {
	// Create request body as JSON
	reqBody := map[string]string{
		"url": opts.URL,
	}
	if opts.Expires != "" {
		reqBody["expires"] = opts.Expires
	}
	if opts.Title != "" {
		reqBody["title"] = opts.Title
	}

	// Encode body as JSON
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(reqBody); err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.BaseURL+"/url", body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set JSON content type
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shortening failed: %s: %s", resp.Status, string(body))
	}

	// Parse JSON response
	var result ShortenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}

func (c *Client) List(opts ListOptions) (*ListResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/urls", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	q := req.URL.Query()
	if opts.Page > 0 {
		q.Add("page", strconv.Itoa(opts.Page))
	}
	if opts.Limit > 0 {
		q.Add("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Sort != "" {
		q.Add("sort", opts.Sort)
	}
	req.URL.RawQuery = q.Encode()

	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list failed: %s: %s", resp.Status, string(body))
	}

	// Parse JSON response
	var result ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}

func (c *Client) Delete(deleteID string) error {
	// Create request
	req, err := http.NewRequest("DELETE", c.BaseURL+"/"+deleteID, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %s: %s", resp.Status, string(body))
	}

	return nil
}

func (c *Client) RequestAPIKey(opts KeyRequestOptions) (*KeyResponse, error) {
	// Create request body
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(map[string]string{
		"email": opts.Email,
		"name":  opts.Name,
	}); err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.BaseURL+"/api-key", body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed: %s: %s", resp.Status, string(body))
	}

	// Parse response
	var result KeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetURLStats(id string) (*URLStatsResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/url/%s/stats", c.BaseURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("getting stats failed: %s: %s", resp.Status, string(body))
	}

	var result URLStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}

func (c *Client) UpdateURLExpiration(id string, expiresIn string) (*ShortenResponse, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(map[string]string{
		"expires_in": expiresIn,
	}); err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/url/%s/expire", c.BaseURL, id), body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("updating expiration failed: %s: %s", resp.Status, string(body))
	}

	var result ShortenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}
