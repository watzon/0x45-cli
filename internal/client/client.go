package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type UploadResponse struct {
	Success   bool   `json:"success"`
	URL       string `json:"url"`
	DeleteURL string `json:"delete_url"`
}

type ShortenResponse struct {
	Success   bool   `json:"success"`
	URL       string `json:"url"`
	DeleteURL string `json:"delete_url"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
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
		Total int `json:"total"`
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"data"`
}

func getBaseURL() string {
	baseURL := viper.GetString("api_url")
	if baseURL == "" {
		baseURL = "https://0x45.st"
	}
	return baseURL
}

func getAPIKey() string {
	return viper.GetString("api_key")
}

func makeRequest(method, path string, query url.Values, body io.Reader) (*http.Response, error) {
	baseURL := getBaseURL()
	apiKey := getAPIKey()

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{}
	return client.Do(req)
}

func UploadFile(filePath string, private bool, expires string) (*UploadResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Prepare query parameters
	query := url.Values{}
	query.Set("filename", filepath.Base(filePath))
	if private {
		query.Set("private", "true")
	}
	if expires != "" {
		query.Set("expires", expires)
	}

	resp, err := makeRequest("POST", "/upload", query, file)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var uploadResp UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, err
	}

	return &uploadResp, nil
}

func ShortenURL(urlToShorten string, private bool, expires string) (*ShortenResponse, error) {
	query := url.Values{}
	query.Set("url", urlToShorten)
	if private {
		query.Set("private", "true")
	}
	if expires != "" {
		query.Set("expires", expires)
	}

	resp, err := makeRequest("POST", "/shorten", query, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var shortenResp ShortenResponse
	if err := json.NewDecoder(resp.Body).Decode(&shortenResp); err != nil {
		return nil, err
	}

	return &shortenResp, nil
}

func Delete(id string) (*DeleteResponse, error) {
	resp, err := makeRequest("DELETE", fmt.Sprintf("/delete/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var deleteResp DeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteResp); err != nil {
		return nil, err
	}

	return &deleteResp, nil
}

func ListPastes(page, limit int) (*ListResponse[PasteListItem], error) {
	query := url.Values{}
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("limit", fmt.Sprintf("%d", limit))

	resp, err := makeRequest("GET", "/pastes", query, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var listResp ListResponse[PasteListItem]
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	return &listResp, nil
}

func ListURLs(page, limit int) (*ListResponse[URLListItem], error) {
	query := url.Values{}
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("limit", fmt.Sprintf("%d", limit))

	resp, err := makeRequest("GET", "/urls", query, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var listResp ListResponse[URLListItem]
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	return &listResp, nil
}
