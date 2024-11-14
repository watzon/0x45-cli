package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	BaseUrl string
	APIKey  string
}

type UploadOptions struct {
	Filename string
	Ext      string
	Expires  string
	Private  bool
}

type UploadResponse struct {
	Success bool `json:"success"`
	Data    struct {
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
	} `json:"data"`
}

type ShortenOptions struct {
	Url     string
	Expires string
	Title   string
}

type ShortenResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Id        string     `json:"id"`
		ShortUrl  string     `json:"short_url"`
		Url       string     `json:"url"`
		Title     string     `json:"title"`
		DeleteUrl string     `json:"delete_url"`
		Clicks    int        `json:"clicks"`
		LastClick *time.Time `json:"last_click"`
		CreatedAt time.Time  `json:"created_at"`
		ExpiresAt *time.Time `json:"expires_at"`
	} `json:"data"`
}

type ListOptions struct {
	Page  int
	Limit int
	Sort  string
}

type PasteListItem struct {
	Id          string     `json:"id"`
	Filename    string     `json:"filename"`
	Size        int64      `json:"size"`
	MimeType    string     `json:"mime_type"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Private     bool       `json:"private"`
	Url         string     `json:"url"`
	RawUrl      string     `json:"raw_url"`
	DownloadUrl string     `json:"download_url"`
	DeleteUrl   string     `json:"delete_url"`
}

type UrlListItem struct {
	Id        string    `json:"id"`
	Url       string    `json:"url"`
	ShortUrl  string    `json:"short_url"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Clicks    int       `json:"clicks"`
	LastClick time.Time `json:"last_click"`
	DeleteUrl string    `json:"delete_url"`
}

type ListResponse[T any] struct {
	Success bool `json:"success"`
	Data    struct {
		Items []T `json:"items"`
		Total int `json:"total"`
		Page  int `json:"page"`
		Limit int `json:"limit"`
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

type UrlStatsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Id        string     `json:"id"`
		Url       string     `json:"url"`
		ShortUrl  string     `json:"short_url"`
		Clicks    int        `json:"clicks"`
		LastClick *time.Time `json:"last_click"`
		CreatedAt time.Time  `json:"created_at"`
		ExpiresAt *time.Time `json:"expires_at"`
	} `json:"data"`
}

type DeleteResponse struct {
	Success bool `json:"success"`
}

func New(baseUrl, apiKey string) *Client {
	return &Client{
		BaseUrl: baseUrl,
		APIKey:  apiKey,
	}
}

func (c *Client) Upload(content io.Reader, query url.Values) (*UploadResponse, error) {
	uploadURL := c.BaseUrl
	if len(query) > 0 {
		uploadURL += "?" + query.Encode()
	}

	req, err := http.NewRequest("POST", uploadURL, content)
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
		return nil, fmt.Errorf("request failed: %s: %s", resp.Status, string(body))
	}

	var result UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}

func (c *Client) Shorten(opts ShortenOptions) (*ShortenResponse, error) {
	reqBody := map[string]string{
		"url": opts.Url,
	}
	if opts.Expires != "" {
		reqBody["expires"] = opts.Expires
	}
	if opts.Title != "" {
		reqBody["title"] = opts.Title
	}

	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(reqBody); err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	var result ShortenResponse
	if err := c.doRequest("POST", "/url", body, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListPastes(opts ListOptions) (*ListResponse[PasteListItem], error) {
	query := make(url.Values)
	if opts.Page > 0 {
		query.Add("page", strconv.Itoa(opts.Page))
	}
	if opts.Limit > 0 {
		query.Add("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Sort != "" {
		query.Add("sort", opts.Sort)
	}

	var result ListResponse[PasteListItem]
	if err := c.doRequest("GET", "/pastes", nil, query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListUrls(opts ListOptions) (*ListResponse[UrlListItem], error) {
	query := make(url.Values)
	if opts.Page > 0 {
		query.Add("page", strconv.Itoa(opts.Page))
	}
	if opts.Limit > 0 {
		query.Add("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Sort != "" {
		query.Add("sort", opts.Sort)
	}

	var result ListResponse[UrlListItem]
	if err := c.doRequest("GET", "/urls", nil, query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Delete(deleteId string) (*DeleteResponse, error) {
	var result DeleteResponse
	if err := c.doRequest("DELETE", "/"+deleteId, nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) RequestAPIKey(opts KeyRequestOptions) (*KeyResponse, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(map[string]string{
		"email": opts.Email,
		"name":  opts.Name,
	}); err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	var result KeyResponse
	if err := c.doRequest("POST", "/api-key", body, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetUrlStats(id string) (*UrlStatsResponse, error) {
	var result UrlStatsResponse
	if err := c.doRequest("GET", fmt.Sprintf("/url/%s/stats", id), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateUrlExpiration(id string, expiresIn string) (*ShortenResponse, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(map[string]string{
		"expires_in": expiresIn,
	}); err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	var result ShortenResponse
	if err := c.doRequest("PUT", fmt.Sprintf("/url/%s/expire", id), body, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Generic helper method for making HTTP requests
func (c *Client) doRequest(method, path string, body io.Reader, query url.Values, result interface{}) error {
	req, err := http.NewRequest(method, c.BaseUrl+path, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if query != nil {
		req.URL.RawQuery = query.Encode()
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed: %s: %s", resp.Status, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
	}

	return nil
}
