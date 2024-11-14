package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			Run:  handleConfigSet,
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
			Run:  handleConfigGet,
		},
		&cobra.Command{
			Use:   "list",
			Short: configCmdStyle.Render("List all config values"),
			Long: lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("List all configuration values"),
				"",
				descriptionStyle.Render("Display all current configuration settings."),
			),
			Run: handleConfigList,
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
			Run:  handleConfigUnset,
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

	// Links subcommand
	linksCmd := &cobra.Command{
		Use:   "urls",
		Short: listCmdStyle.Render("List your shortened URLs"),
		RunE:  handleListUrls,
	}

	// Pastes subcommand
	pastesCmd := &cobra.Command{
		Use:   "pastes",
		Short: listCmdStyle.Render("List your uploaded pastes"),
		RunE:  handleListPastes,
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
		RunE: handleUpload,
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
		RunE: handleShorten,
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
		RunE: handleDelete,
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
		RunE: handleKeyRequest,
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
		Run:   handleKeyStatus,
	}

	cmd.AddCommand(requestCmd, statusCmd)
	return cmd
}
