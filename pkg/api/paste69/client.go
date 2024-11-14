package paste69

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

type UploadRequest struct {
	File     string `json:"file"`
	Private  bool   `json:"private,omitempty"`
	Expires  string `json:"expires,omitempty"`
	Filename string `json:"filename,omitempty"`
}

type ShortenRequest struct {
	URL     string `json:"url"`
	Private bool   `json:"private,omitempty"`
	Expires string `json:"expires,omitempty"`
}

type UploadResponse struct {
	Success   bool   `json:"success"`
	URL       string `json:"url,omitempty"`
	DeleteURL string `json:"delete_url,omitempty"`
	Error     string `json:"error,omitempty"`
}

type ShortenResponse struct {
	Success   bool   `json:"success"`
	URL       string `json:"url,omitempty"`
	DeleteURL string `json:"delete_url,omitempty"`
	Error     string `json:"error,omitempty"`
}

type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type PasteListItem struct {
	Id        string `json:"id"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
	URL       string `json:"url"`
}

type URLListItem struct {
	Id          string `json:"id"`
	URL         string `json:"url"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
}

type ListResponse[T any] struct {
	Success bool `json:"success"`
	Data    struct {
		Items []T `json:"items"`
	} `json:"data"`
	Error string `json:"error,omitempty"`
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) Upload(filePath string, private bool, expires string) (*UploadResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %w", err)
	}

	params := url.Values{}
	if private {
		params.Set("private", "true")
	}
	if expires != "" {
		params.Set("expires", expires)
	}

	reqURL := fmt.Sprintf("%s/upload?%s", c.BaseURL, params.Encode())
	req, err := http.NewRequest("POST", reqURL, file)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("X-Filename", filepath.Base(filePath))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}

func (c *Client) Shorten(targetURL string, private bool, expires string) (*ShortenResponse, error) {
	params := url.Values{}
	if private {
		params.Set("private", "true")
	}
	if expires != "" {
		params.Set("expires", expires)
	}

	reqURL := fmt.Sprintf("%s/shorten?%s", c.BaseURL, params.Encode())
	body := strings.NewReader(targetURL)
	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result ShortenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}

func (c *Client) Delete(id string) (*GenericResponse, error) {
	reqURL := fmt.Sprintf("%s/delete/%s", c.BaseURL, id)
	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result GenericResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}

func (c *Client) ListPastes(page, perPage int) (*ListResponse[PasteListItem], error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("per_page", fmt.Sprintf("%d", perPage))

	reqURL := fmt.Sprintf("%s/pastes?%s", c.BaseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result ListResponse[PasteListItem]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}

func (c *Client) ListURLs(page, perPage int) (*ListResponse[URLListItem], error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("per_page", fmt.Sprintf("%d", perPage))

	reqURL := fmt.Sprintf("%s/urls?%s", c.BaseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result ListResponse[URLListItem]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}
