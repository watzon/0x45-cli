package main

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/h2non/filetype"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func handleConfigSet(cmd *cobra.Command, args []string) {
	key, value := args[0], args[1]
	viper.Set(key, value)
	if err := viper.WriteConfig(); err != nil {
		if err := viper.SafeWriteConfig(); err != nil {
			cobra.CheckErr(err)
		}
	}
	fmt.Printf("%s %s to %s\n",
		successStyle.Render("✓"),
		titleStyle.Render(key),
		subtitleStyle.Render(value))
}

func handleConfigGet(cmd *cobra.Command, args []string) {
	key := args[0]
	value := viper.Get(key)
	if value == nil {
		fmt.Printf("Config key '%s' not found\n", key)
		return
	}
	fmt.Printf("%v\n", value)
}

func handleConfigList(cmd *cobra.Command, args []string) {
	fmt.Printf("\n%s\n\n", titleStyle.Render("Current Configuration"))

	settings := viper.AllSettings()
	var output []string

	for key, value := range settings {
		output = append(output,
			formatKeyValue(key, fmt.Sprintf("%v", value)),
		)
	}

	fmt.Println(lipgloss.JoinVertical(lipgloss.Left, output...))
	fmt.Println()
}

func handleConfigUnset(cmd *cobra.Command, args []string) {
	key := args[0]
	if !viper.IsSet(key) {
		fmt.Printf("%s Config key '%s' not found\n",
			errorStyle.Render("✗"),
			key)
		return
	}

	viper.Set(key, nil)
	if err := viper.WriteConfig(); err != nil {
		cobra.CheckErr(err)
	}

	fmt.Printf("%s Removed config key %s\n",
		successStyle.Render("✓"),
		titleStyle.Render(key))
}

func handleListUrls(cmd *cobra.Command, args []string) error {
	if err := validateAPIKey(); err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	page, _ := cmd.Flags().GetInt("page")
	sort, _ := cmd.Flags().GetString("sort")

	c := New(
		viper.GetString("api_url"),
		viper.GetString("api_key"),
	)

	resp, err := c.ListUrls(ListOptions{
		Limit: limit,
		Page:  page,
		Sort:  sort,
	})
	if err != nil {
		return err
	}

	if len(resp.Data.Items) == 0 {
		fmt.Println(descriptionStyle.Render("No shortened URLs found"))
		return nil
	}

	fmt.Printf("\n%s\n\n", titleStyle.Render("Your Shortened URLs"))

	for _, item := range resp.Data.Items {
		fmt.Println(formatUrlEntry(item))
	}

	fmt.Printf("%s\n\n",
		subtitleStyle.Render(fmt.Sprintf(
			"Page %d of %d (showing %d of %d total)",
			resp.Data.Page,
			(resp.Data.Total+resp.Data.Limit-1)/resp.Data.Limit,
			len(resp.Data.Items),
			resp.Data.Total,
		)))

	return nil
}

func handleListPastes(cmd *cobra.Command, args []string) error {
	if err := validateAPIKey(); err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	page, _ := cmd.Flags().GetInt("page")
	sort, _ := cmd.Flags().GetString("sort")

	c := New(
		viper.GetString("api_url"),
		viper.GetString("api_key"),
	)

	resp, err := c.ListPastes(ListOptions{
		Limit: limit,
		Page:  page,
		Sort:  sort,
	})
	if err != nil {
		return err
	}

	if len(resp.Data.Items) == 0 {
		fmt.Println(descriptionStyle.Render("No uploaded pastes found"))
		return nil
	}

	fmt.Printf("\n%s\n\n", titleStyle.Render("Your Uploaded Pastes"))

	for _, item := range resp.Data.Items {
		fmt.Println(formatPasteEntry(item))
	}

	fmt.Printf("%s\n\n",
		subtitleStyle.Render(fmt.Sprintf(
			"Page %d of %d (showing %d of %d total)",
			resp.Data.Page,
			(resp.Data.Total+resp.Data.Limit-1)/resp.Data.Limit,
			len(resp.Data.Items),
			resp.Data.Total,
		)))

	return nil
}

func handleUpload(cmd *cobra.Command, args []string) error {
	expires, _ := cmd.Flags().GetString("expires")
	private, _ := cmd.Flags().GetBool("private")
	customFilename, _ := cmd.Flags().GetString("filename")
	customExt, _ := cmd.Flags().GetString("ext")

	if private {
		if err := validateAPIKey(); err != nil {
			return fmt.Errorf("private uploads require an API key: %w", err)
		}
	}

	if expires != "" {
		duration, err := time.ParseDuration(expires)
		if err != nil {
			return fmt.Errorf("invalid expiry duration: %w", err)
		}

		hasAPIKey := viper.GetString("api_key") != ""
		maxDays := 730
		if !hasAPIKey {
			maxDays = 128
		}
		maxDuration := time.Duration(maxDays) * 24 * time.Hour

		if duration > maxDuration {
			return fmt.Errorf("%s maximum expiry without API key is %d days",
				errorStyle.Render(""),
				maxDays)
		}
	}

	c := New(
		viper.GetString("api_url"),
		viper.GetString("api_key"),
	)

	query := url.Values{}
	if expires != "" {
		query.Set("expires", expires)
	}
	if private {
		query.Set("private", "true")
	}

	var fileContent []byte
	var err error

	if len(args) > 0 {
		fileContent, err = os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		if customFilename == "" {
			query.Set("filename", filepath.Base(args[0]))
		}
		if customExt == "" && filepath.Ext(args[0]) != "" {
			query.Set("ext", filepath.Ext(args[0])[1:])
		}
	} else {
		fileContent, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		// Try to detect file type from content
		kind, err := filetype.Match(fileContent)
		if err == nil && kind != filetype.Unknown {
			if customFilename == "" {
				query.Set("filename", fmt.Sprintf("upload.%s", kind.Extension))
			}
			if customExt == "" {
				query.Set("ext", kind.Extension)
			}
		} else {
			// If we can't detect a specific binary format, assume it's text
			if customFilename == "" {
				query.Set("filename", "paste.txt")
			}
			if customExt == "" {
				query.Set("ext", "txt")
			}
		}
	}

	if customFilename != "" {
		query.Set("filename", customFilename)
	}
	if customExt != "" {
		query.Set("ext", customExt)
	}

	resp, err := c.Upload(bytes.NewReader(fileContent), query)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s %s\n\n",
		successStyle.Render("✓"),
		titleStyle.Render("Upload successful!"))

	output := formatUploadResponse(resp)
	fmt.Println(output)
	fmt.Println()

	return nil
}

func handleShorten(cmd *cobra.Command, args []string) error {
	if err := validateAPIKey(); err != nil {
		return err
	}

	url := args[0]
	expires, _ := cmd.Flags().GetString("expires")
	title, _ := cmd.Flags().GetString("title")

	c := New(
		viper.GetString("api_url"),
		viper.GetString("api_key"),
	)

	resp, err := c.Shorten(ShortenOptions{
		Url:     url,
		Expires: expires,
		Title:   title,
	})
	if err != nil {
		return err
	}

	fmt.Printf("\n%s %s\n\n",
		successStyle.Render("✓"),
		titleStyle.Render("URL shortened successfully!"))

	output := formatShortenResponse(resp)
	fmt.Println(output)
	fmt.Println()

	return nil
}

func handleDelete(cmd *cobra.Command, args []string) error {
	if err := validateAPIKey(); err != nil {
		return err
	}

	deleteId := args[0]

	c := New(
		viper.GetString("api_url"),
		viper.GetString("api_key"),
	)

	if err := c.Delete(deleteId); err != nil {
		return err
	}

	fmt.Printf("\n%s %s\n\n",
		successStyle.Render("✓"),
		titleStyle.Render("Content deleted successfully!"))

	return nil
}

func handleKeyRequest(cmd *cobra.Command, args []string) error {
	email, _ := cmd.Flags().GetString("email")
	name, _ := cmd.Flags().GetString("name")

	if email == "" || name == "" {
		return fmt.Errorf("email and name are required")
	}

	c := New(
		viper.GetString("api_url"),
		"",
	)

	resp, err := c.RequestAPIKey(KeyRequestOptions{
		Email: email,
		Name:  name,
	})
	if err != nil {
		return err
	}

	fmt.Printf("\n%s %s\n\n",
		successStyle.Render("✓"),
		titleStyle.Render(resp.Message))

	return nil
}

func handleKeyStatus(cmd *cobra.Command, args []string) {
	fmt.Println()

	if apiKey := viper.GetString("api_key"); apiKey != "" {
		output := lipgloss.JoinVertical(lipgloss.Left,
			fmt.Sprintf("%s %s",
				successStyle.Render("✓"),
				titleStyle.Render("API Key Configuration")),
			"",
			formatKeyValue("API Key", apiKey),
			formatKeyValue("Max Expiry", "730 days (2 years)"),
			formatKeyValue("Private Pastes", "Enabled"),
		)
		fmt.Println(output)
	} else {
		output := lipgloss.JoinVertical(lipgloss.Left,
			fmt.Sprintf("%s %s",
				errorStyle.Render("✗"),
				titleStyle.Render("No API key configured")),
			"",
			descriptionStyle.Render(fmt.Sprintf(
				"Run %s to request a key",
				keyCmdStyle.Render("0x45 key request --email you@example.com --name \"Your Name\""))),
		)
		fmt.Println(output)
	}
	fmt.Println()
}

// Helper functions for formatting responses
func formatUploadResponse(resp *UploadResponse) string {
	output := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(resp.Data.Filename),
		urlStyle.Render(resp.Data.Url),
		formatKeyValue("Created", resp.Data.CreatedAt.Format("2006-01-02")),
	)

	if resp.Data.ExpiresAt != nil {
		output = lipgloss.JoinVertical(lipgloss.Left,
			output,
			formatKeyValue("Expires", resp.Data.ExpiresAt.Format("2006-01-02")),
		)
	}

	output = lipgloss.JoinVertical(lipgloss.Left,
		output,
		formatKeyValue("Size", humanize.Bytes(uint64(resp.Data.Size))),
		formatKeyValue("ID", resp.Data.Id),
		"",
		subtitleStyle.Render("Additional URLs:"),
		formatKeyValue("Raw", urlStyle.Render(resp.Data.RawUrl)),
		formatKeyValue("Download", urlStyle.Render(resp.Data.DownloadUrl)),
		formatKeyValue("Delete", urlStyle.Render(resp.Data.DeleteUrl)),
	)

	return output
}

func formatShortenResponse(resp *ShortenResponse) string {
	output := lipgloss.JoinVertical(lipgloss.Left,
		urlStyle.Render(resp.Data.ShortUrl),
		subtitleStyle.Render(fmt.Sprintf("→ %s", resp.Data.Url)),
		formatKeyValue("Created", resp.Data.CreatedAt.Format("2006-01-02")),
		formatKeyValue("Clicks", strconv.Itoa(resp.Data.Clicks)),
		formatKeyValue("ID", resp.Data.Id),
		"",
		formatKeyValue("Delete", urlStyle.Render(resp.Data.DeleteUrl)),
	)

	if resp.Data.ExpiresAt != nil {
		output = lipgloss.JoinVertical(lipgloss.Left,
			output,
			"",
			formatKeyValue("Expires", resp.Data.ExpiresAt.Format("2006-01-02")),
		)
	}

	return output
}

func formatUrlEntry(item UrlListItem) string {
	return lipgloss.JoinVertical(lipgloss.Left,
		urlStyle.Render(item.ShortUrl),
		subtitleStyle.Render(fmt.Sprintf("→ %s", item.Url)),
		descriptionStyle.Render(fmt.Sprintf(
			"Created: %s • Expires: %s • Clicks: %d • ID: %s",
			item.CreatedAt.Format("2006-01-02"),
			item.ExpiresAt.Format("2006-01-02"),
			item.Clicks,
			item.Id,
		)),
		"",
	)
}

func formatPasteEntry(item PasteListItem) string {
	size := "-"
	if item.Size > 0 {
		size = humanize.Bytes(uint64(item.Size))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(item.Filename),
		urlStyle.Render(item.Url),
		descriptionStyle.Render(fmt.Sprintf(
			"Created: %s • Expires: %s  Size: %s • ID: %s",
			item.CreatedAt.Format("2006-01-02"),
			item.ExpiresAt.Format("2006-01-02"),
			size,
			item.Id,
		)),
		"",
	)
}
