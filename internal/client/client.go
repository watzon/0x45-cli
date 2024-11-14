package client

import (
	"github.com/spf13/viper"
	"github.com/watzon/0x45-cli/pkg/api/paste69"
)

var client *paste69.Client

func Initialize() {
	client = paste69.NewClient(
		viper.GetString("api_url"),
		viper.GetString("api_key"),
	)
}

func init() {
	Initialize()
}

func UploadFile(filePath string, private bool, expires string) (*paste69.UploadResponse, error) {
	return client.Upload(filePath, private, expires)
}

func ShortenURL(url string, private bool, expires string) (*paste69.ShortenResponse, error) {
	return client.Shorten(url, private, expires)
}

func Delete(id string) (*paste69.GenericResponse, error) {
	return client.Delete(id)
}

func ListPastes(page, perPage int) (*paste69.ListResponse[paste69.PasteListItem], error) {
	return client.ListPastes(page, perPage)
}

func ListURLs(page, perPage int) (*paste69.ListResponse[paste69.URLListItem], error) {
	return client.ListURLs(page, perPage)
}
