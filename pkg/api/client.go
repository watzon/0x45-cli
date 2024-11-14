package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client represents a generic API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new API client instance
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

// ListResponse is a generic response type for paginated lists
type ListResponse[T any] struct {
	Success bool `json:"success"`
	Data    struct {
		Items []T `json:"items"`
		Total int `json:"total"`
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"data"`
}

// GenericResponse represents a basic API response
type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// makeRequest is a generic method to make HTTP requests
func (c *Client) MakeRequest(method, path string, query url.Values, body io.Reader) (*http.Response, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	u.Path = path
	if query != nil {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// DecodeResponse is a helper function to decode JSON responses
func DecodeResponse[T any](resp *http.Response) (*T, error) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
