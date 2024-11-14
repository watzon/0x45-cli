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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dustin/go-humanize"
)

var (
	// Config file paths
	cfgFile string

	// Style definitions
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ECF0F1"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BDC3C7"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2ECC71"))

	urlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3498DB")).
			Underline(true)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E67E22"))

	// Command category colors
	uploadCmdStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2ECC71")) // Green for upload/paste commands

	urlCmdStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3498DB")) // Blue for URL-related commands

	configCmdStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E67E22")) // Orange for config commands

	listCmdStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#9B59B6")) // Purple for list/view commands

	deleteCmdStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E74C3C")) // Red for delete/remove commands

	// Help styles
	exampleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#27AE60"))

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ECF0F1"))

	// Help text styles
	flagNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F1C40F")) // Yellow for flag names

	flagDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BDC3C7")) // Light gray for flag descriptions

	usageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3498DB")) // Blue for usage text

	keyCmdStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F1C40F")) // Gold/yellow for auth/key commands

	// Add a specific style for keys in key-value pairs
	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#9B59B6")) // Purple for keys
)

// Helper function for formatting key-value pairs
func formatKeyValue(key, value string) string {
	return fmt.Sprintf("%s: %s",
		keyStyle.Render(key),
		descriptionStyle.Render(value))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search for config in home directory with name ".0x45" (with any supported extension)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".0x45")

		// Ensure config directory exists
		configDir := filepath.Join(home, ".config", "0x45")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			cobra.CheckErr(err)
		}
	}

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			cobra.CheckErr(err)
		}
	}

	// Set defaults
	viper.SetDefault("api_url", "https://0x45.st")

	// Bind flags to viper
	viper.BindEnv("api_key", "OX45_API_KEY")
}

func validateAPIKey() error {
	apiKey := viper.GetString("api_key")
	if apiKey == "" {
		return fmt.Errorf("%s API key required. Set it with: %s",
			errorStyle.Render("✗"),
			configCmdStyle.Render("0x45 config set api_key <your-key>"))
	}
	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "0x45",
		Short: titleStyle.Render("A CLI tool for interacting with 0x45.st paste service"),
		Long: lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("0x45.st CLI Tool"),
			"",
			descriptionStyle.Render("A beautiful command-line interface for managing pastes, files, and URLs on 0x45.st."),
			"",
			exampleStyle.Render("Examples:"),
			fmt.Sprintf("  %s file.txt", uploadCmdStyle.Render("0x45 upload")),
			fmt.Sprintf("  cat image.png | %s", uploadCmdStyle.Render("0x45 upload")),
			fmt.Sprintf("  %s https://example.com", urlCmdStyle.Render("0x45 shorten")),
		),
	}

	// Global flags with styled descriptions
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		flagDescStyle.Render("config file (default is $HOME/.0x45.yaml)"))
	rootCmd.PersistentFlags().String("api-key", "",
		flagDescStyle.Render("API key for authentication"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))

	// Initialize config
	cobra.OnInitialize(initConfig)

	// Add commands
	rootCmd.AddCommand(
		newConfigCommand(),
		newListCommand(),
		newUploadCommand(),
		newShortenCommand(),
		newDeleteCommand(),
		newKeyCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(errorStyle.Render(err.Error()))
		os.Exit(1)
	}
}

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: configCmdStyle.Render("Manage configuration settings"),
		Long: lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Manage your 0x45.st configuration"),
			"",
			descriptionStyle.Render("View and modify configuration settings such as API keys,"),
			descriptionStyle.Render("default expiration times, and API endpoints."),
			"",
			fmt.Sprintf("%s:", usageStyle.Render("Usage")),
			fmt.Sprintf("  %s <command>", configCmdStyle.Render("0x45 config")),
			"",
			fmt.Sprintf("%s:", usageStyle.Render("Commands")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("set <key> <value>"),
				flagDescStyle.Render("Set a configuration value")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("get <key>"),
				flagDescStyle.Render("Get a configuration value")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("list"),
				flagDescStyle.Render("List all configuration values")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("unset [key]"),
				flagDescStyle.Render("Unset a configuration value")),
			"",
			exampleStyle.Render("Examples:"),
			fmt.Sprintf("  %s set api_key your-key", configCmdStyle.Render("0x45 config")),
			fmt.Sprintf("  %s get api_key", configCmdStyle.Render("0x45 config")),
			fmt.Sprintf("  %s list", configCmdStyle.Render("0x45 config")),
			fmt.Sprintf("  %s unset default_expiry", configCmdStyle.Render("0x45 config")),
		),
	}

	// Add subcommands with styled help
	cmd.AddCommand(
		&cobra.Command{
			Use:   "set [key] [value]",
			Short: configCmdStyle.Render("Set a config value"),
			Long: lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("Set a configuration value"),
				"",
				descriptionStyle.Render("Set a configuration value that will be saved to your config file."),
				"",
				fmt.Sprintf("%s:", usageStyle.Render("Usage")),
				fmt.Sprintf("  %s set [key] [value]", configCmdStyle.Render("0x45 config")),
				"",
				exampleStyle.Render("Examples:"),
				fmt.Sprintf("  %s set api_key your-key", configCmdStyle.Render("0x45 config")),
				fmt.Sprintf("  %s set default_expiry 30d", configCmdStyle.Render("0x45 config")),
			),
			Args: cobra.ExactArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
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
			},
		},
		&cobra.Command{
			Use:   "get [key]",
			Short: configCmdStyle.Render("Get a config value"),
			Long: lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("Get a configuration value"),
				"",
				descriptionStyle.Render("Retrieve the current value of a configuration setting."),
				"",
				fmt.Sprintf("%s:", usageStyle.Render("Usage")),
				fmt.Sprintf("  %s get [key]", configCmdStyle.Render("0x45 config")),
				"",
				exampleStyle.Render("Examples:"),
				fmt.Sprintf("  %s get api_key", configCmdStyle.Render("0x45 config")),
				fmt.Sprintf("  %s get default_expiry", configCmdStyle.Render("0x45 config")),
			),
			Args: cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				key := args[0]
				value := viper.Get(key)
				if value == nil {
					fmt.Printf("Config key '%s' not found\n", key)
					return
				}
				fmt.Printf("%v\n", value)
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: configCmdStyle.Render("List all config values"),
			Long: lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("List all configuration values"),
				"",
				descriptionStyle.Render("Display all current configuration settings."),
			),
			Run: func(cmd *cobra.Command, args []string) {
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
			},
		},
		&cobra.Command{
			Use:   "unset [key]",
			Short: configCmdStyle.Render("Unset a config value"),
			Long: lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("Unset a configuration value"),
				"",
				descriptionStyle.Render("Remove a configuration value from your config file."),
				"",
				fmt.Sprintf("%s:", usageStyle.Render("Usage")),
				fmt.Sprintf("  %s unset [key]", configCmdStyle.Render("0x45 config")),
				"",
				exampleStyle.Render("Examples:"),
				fmt.Sprintf("  %s unset default_expiry", configCmdStyle.Render("0x45 config")),
			),
			Args: cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
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
			},
		},
	)

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: listCmdStyle.Render("List your pastes and shortened URLs"),
		Long: lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("List your pastes and shortened URLs"),
			"",
			descriptionStyle.Render("Show a record of either your uploaded pastes or shortened URLs."),
			"",
			fmt.Sprintf("%s:", usageStyle.Render("Usage")),
			fmt.Sprintf("  %s <command> [flags]", listCmdStyle.Render("0x45 list")),
			"",
			fmt.Sprintf("%s:", usageStyle.Render("Commands")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("pastes"),
				flagDescStyle.Render("List your uploaded pastes")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("urls"),
				flagDescStyle.Render("List your shortened URLs")),
		),
	}

	// Common flags for both subcommands
	var limit int
	var page int
	var sort string

	// Helper function to format a URL entry
	formatUrlEntry := func(item UrlListItem) string {
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
			"", // Empty line for spacing
		)
	}

	// Helper function to format a paste entry
	formatPasteEntry := func(item PasteListItem) string {
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
			"", // Empty line for spacing
		)
	}

	// Links subcommand
	linksCmd := &cobra.Command{
		Use:   "urls",
		Short: listCmdStyle.Render("List your shortened URLs"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateAPIKey(); err != nil {
				return err
			}

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

			// Print header
			fmt.Printf("\n%s\n\n", titleStyle.Render("Your Shortened URLs"))

			// Print each URL entry
			for _, item := range resp.Data.Items {
				fmt.Println(formatUrlEntry(item))
			}

			// Print pagination info
			fmt.Printf("%s\n\n",
				subtitleStyle.Render(fmt.Sprintf(
					"Page %d of %d (showing %d of %d total)",
					resp.Data.Page,
					(resp.Data.Total+resp.Data.Limit-1)/resp.Data.Limit,
					len(resp.Data.Items),
					resp.Data.Total,
				)))

			return nil
		},
	}

	// Pastes subcommand
	pastesCmd := &cobra.Command{
		Use:   "pastes",
		Short: listCmdStyle.Render("List your uploaded pastes"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateAPIKey(); err != nil {
				return err
			}

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

			// Print header
			fmt.Printf("\n%s\n\n", titleStyle.Render("Your Uploaded Pastes"))

			// Print each paste entry
			for _, item := range resp.Data.Items {
				fmt.Println(formatPasteEntry(item))
			}

			// Print pagination info
			fmt.Printf("%s\n\n",
				subtitleStyle.Render(fmt.Sprintf(
					"Page %d of %d (showing %d of %d total)",
					resp.Data.Page,
					(resp.Data.Total+resp.Data.Limit-1)/resp.Data.Limit,
					len(resp.Data.Items),
					resp.Data.Total,
				)))

			return nil
		},
	}

	cmd.AddCommand(pastesCmd, linksCmd)

	// Set flags for subcommands
	pastesCmd.Flags().IntVarP(&limit, "limit", "l", 10, flagDescStyle.Render("Limit the number of results"))
	pastesCmd.Flags().IntVarP(&page, "page", "p", 1, flagDescStyle.Render("Page number"))
	pastesCmd.Flags().StringVarP(&sort, "sort", "s", "created_at", flagDescStyle.Render("Sort by created_at, expires_at, or clicks"))

	linksCmd.Flags().IntVarP(&limit, "limit", "l", 10, flagDescStyle.Render("Limit the number of results"))
	linksCmd.Flags().IntVarP(&page, "page", "p", 1, flagDescStyle.Render("Page number"))
	linksCmd.Flags().StringVarP(&sort, "sort", "s", "created_at", flagDescStyle.Render("Sort by created_at, expires_at, or clicks"))

	return cmd
}

func newUploadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload [file]",
		Short: uploadCmdStyle.Render("Upload a file or paste content"),
		Long: lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Upload files or text content to 0x45.st"),
			"",
			descriptionStyle.Render("Upload files directly or pipe content through stdin."),
			descriptionStyle.Render("Supports various expiration times and private uploads."),
			"",
			fmt.Sprintf("%s:", usageStyle.Render("Usage")),
			fmt.Sprintf("  %s [flags] [file]", uploadCmdStyle.Render("0x45 upload")),
			"",
			fmt.Sprintf("%s:", usageStyle.Render("Flags")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("-e, --expires <duration>"),
				flagDescStyle.Render("Expiration time (e.g., 24h, 7d)")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("-p, --private"),
				flagDescStyle.Render("Make the paste private")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("-f, --filename <filename>"),
				flagDescStyle.Render("Override the filename")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("-x, --ext <ext>"),
				flagDescStyle.Render("Override the file extension")),
			"",
			exampleStyle.Render("Examples:"),
			fmt.Sprintf("  %s file.txt", uploadCmdStyle.Render("0x45 upload")),
			fmt.Sprintf("  %s --expires 24h --private screenshot.png", uploadCmdStyle.Render("0x45 upload")),
			fmt.Sprintf("  cat image.png | %s", uploadCmdStyle.Render("0x45 upload")),
		),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			expires, _ := cmd.Flags().GetString("expires")
			private, _ := cmd.Flags().GetBool("private")
			customFilename, _ := cmd.Flags().GetString("filename")
			customExt, _ := cmd.Flags().GetString("ext")

			// Validate private flag requires API key
			if private {
				if err := validateAPIKey(); err != nil {
					return fmt.Errorf("private uploads require an API key: %w", err)
				}
			}

			// Validate expiry duration based on API key presence
			if expires != "" {
				duration, err := time.ParseDuration(expires)
				if err != nil {
					return fmt.Errorf("invalid expiry duration: %w", err)
				}

				// Check limits based on API key presence
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

			// Create client
			c := New(
				viper.GetString("api_url"),
				viper.GetString("api_key"),
			)

			// Build query parameters
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
				// Read from file
				fileContent, err = os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("reading file: %w", err)
				}

				// Get filename and extension from the actual file if not overridden
				if customFilename == "" {
					query.Set("filename", filepath.Base(args[0]))
				}
				if customExt == "" && filepath.Ext(args[0]) != "" {
					query.Set("ext", filepath.Ext(args[0])[1:]) // Remove the dot
				}
			} else {
				// Read from stdin
				fileContent, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}

				if customFilename == "" {
					query.Set("filename", "paste.txt")
				}
				if customExt == "" {
					query.Set("ext", "txt")
				}
			}

			// Add overrides if specified
			if customFilename != "" {
				query.Set("filename", customFilename)
			}
			if customExt != "" {
				query.Set("ext", customExt)
			}

			// Upload content with raw body
			resp, err := c.Upload(bytes.NewReader(fileContent), query)
			if err != nil {
				return err
			}

			fmt.Printf("\n%s %s\n\n",
				successStyle.Render("✓"),
				titleStyle.Render("Upload successful!"))

			// Format the response similar to list command
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

			fmt.Println(output)
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringP("expires", "e", viper.GetString("default_expiry"),
		flagDescStyle.Render("Expiration time (e.g., 24h, 7d)"))
	cmd.Flags().BoolP("private", "p", false,
		flagDescStyle.Render("Make the paste private"))
	cmd.Flags().StringP("filename", "f", "",
		flagDescStyle.Render("Override the filename"))
	cmd.Flags().StringP("ext", "x", "",
		flagDescStyle.Render("Override the file extension"))
	return cmd
}

func newShortenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shorten [url]",
		Short: urlCmdStyle.Render("Shorten a URL"),
		Long: lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Create shortened URLs using 0x45.st"),
			"",
			descriptionStyle.Render("Shorten long URLs into memorable short links."),
			descriptionStyle.Render("Optionally add titles and set expiration times."),
			"",
			fmt.Sprintf("%s:", usageStyle.Render("Usage")),
			fmt.Sprintf("  %s [flags] <url>", urlCmdStyle.Render("0x45 shorten")),
			"",
			fmt.Sprintf("%s:", usageStyle.Render("Flags")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("-e, --expires <duration>"),
				flagDescStyle.Render("Expiration time (e.g., 24h, 7d)")),
			fmt.Sprintf("  %s  %s",
				flagNameStyle.Render("-t, --title <title>"),
				flagDescStyle.Render("URL title")),
			"",
			exampleStyle.Render("Examples:"),
			fmt.Sprintf("  %s https://example.com", urlCmdStyle.Render("0x45 shorten")),
			fmt.Sprintf("  %s --title 'My Site' --expires 30d https://example.com", urlCmdStyle.Render("0x45 shorten")),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate API key first
			if err := validateAPIKey(); err != nil {
				return err
			}

			url := args[0]
			expires, _ := cmd.Flags().GetString("expires")
			title, _ := cmd.Flags().GetString("title")

			// Create client
			c := New(
				viper.GetString("api_url"),
				viper.GetString("api_key"),
			)

			// Shorten URL
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

			fmt.Println(output)
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringP("expires", "e", viper.GetString("default_expiry"),
		flagDescStyle.Render("Expiration time (e.g., 24h, 7d)"))
	cmd.Flags().StringP("title", "t", "",
		flagDescStyle.Render("URL title"))
	return cmd
}

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [delete-id]",
		Short: deleteCmdStyle.Render("Delete a paste or shortened URL"),
		Long: lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Delete a paste or shortened URL"),
			"",
			descriptionStyle.Render("Remove content using its delete ID."),
			descriptionStyle.Render("The delete ID is provided when content is created."),
			"",
			fmt.Sprintf("%s:", usageStyle.Render("Usage")),
			fmt.Sprintf("  %s <delete-id>", deleteCmdStyle.Render("0x45 delete")),
			"",
			exampleStyle.Render("Examples:"),
			fmt.Sprintf("  %s abc123", deleteCmdStyle.Render("0x45 delete")),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate API key first
			if err := validateAPIKey(); err != nil {
				return err
			}

			deleteId := args[0]

			// Create client
			c := New(
				viper.GetString("api_url"),
				viper.GetString("api_key"),
			)

			// Delete content
			if err := c.Delete(deleteId); err != nil {
				return err
			}

			fmt.Printf("\n%s %s\n\n",
				successStyle.Render("✓"),
				titleStyle.Render("Content deleted successfully!"))

			return nil
		},
	}

	return cmd
}

func newKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: keyCmdStyle.Render("Manage API keys"),
		Long: lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Manage your 0x45.st API key"),
			"",
			descriptionStyle.Render("API keys provide additional features like:"),
			descriptionStyle.Render("• Private pastes"),
			descriptionStyle.Render("• Longer expiration times (up to 2 years)"),
			descriptionStyle.Render("• List and manage your uploads"),
		),
	}

	// Add request subcommand
	requestCmd := &cobra.Command{
		Use:   "request",
		Short: keyCmdStyle.Render("Request a new API key"),
		Long: lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Request a new API key"),
			"",
			descriptionStyle.Render("Request an API key by providing your email and name."),
			descriptionStyle.Render("You'll receive a verification email to activate your key."),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			name, _ := cmd.Flags().GetString("name")

			if email == "" || name == "" {
				return fmt.Errorf("email and name are required")
			}

			// Create client
			c := New(
				viper.GetString("api_url"),
				"", // No API key needed for this request
			)

			// Request key
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
		},
	}

	// Add flags for request command
	requestCmd.Flags().String("email", "", flagDescStyle.Render("Your email address"))
	requestCmd.Flags().String("name", "", flagDescStyle.Render("Your name"))
	requestCmd.MarkFlagRequired("email")
	requestCmd.MarkFlagRequired("name")

	// Add status subcommand (shows current key info)
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: keyCmdStyle.Render("Show API key status"),
		Run: func(cmd *cobra.Command, args []string) {
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
		},
	}

	cmd.AddCommand(requestCmd, statusCmd)
	return cmd
}
