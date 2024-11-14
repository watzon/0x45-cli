package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/watzon/0x45-cli/internal/client"
	"github.com/watzon/0x45-cli/internal/theme"
)

func NewUploadCmd() *cobra.Command {
	var private bool
	var expires string

	cmd := &cobra.Command{
		Use:   "upload [file]",
		Short: "Upload a file to 0x45.st",
		Args:  cobra.ExactArgs(1),
		RunE:  Upload,
	}

	cmd.Flags().BoolVar(&private, "private", false, "Make the upload private")
	cmd.Flags().StringVar(&expires, "expires", "", "Set expiration time (e.g. 24h)")

	return cmd
}

func Upload(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	filePath := args[0]
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	private, err := cmd.Flags().GetBool("private")
	if err != nil {
		return err
	}

	expires, err := cmd.Flags().GetString("expires")
	if err != nil {
		return err
	}

	resp, err := client.UploadFile(filePath, private, expires)
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("error uploading file: %s", resp.Error)
	}

	fmt.Fprintln(cmd.OutOrStdout(), resp.URL)
	if resp.DeleteURL != "" {
		fmt.Fprintln(cmd.OutOrStdout(), "Delete URL:", resp.DeleteURL)
	}

	return nil
}

func NewShortenCmd() *cobra.Command {
	var private bool
	var expires string

	cmd := &cobra.Command{
		Use:   "shorten [url]",
		Short: "Shorten a URL using 0x45.st",
		Args:  cobra.ExactArgs(1),
		RunE:  Shorten,
	}

	cmd.Flags().BoolVar(&private, "private", false, "Make the URL private")
	cmd.Flags().StringVar(&expires, "expires", "", "Set expiration time (e.g. 24h)")

	return cmd
}

func Shorten(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	private, err := cmd.Flags().GetBool("private")
	if err != nil {
		return err
	}

	expires, err := cmd.Flags().GetString("expires")
	if err != nil {
		return err
	}

	resp, err := client.ShortenURL(args[0], private, expires)
	if err != nil {
		return fmt.Errorf("error shortening URL: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("error shortening URL: %s", resp.Error)
	}

	fmt.Fprintln(cmd.OutOrStdout(), resp.URL)
	if resp.DeleteURL != "" {
		fmt.Fprintln(cmd.OutOrStdout(), "Delete URL:", resp.DeleteURL)
	}

	return nil
}

func NewListCmd() *cobra.Command {
	var page int
	var limit int

	cmd := &cobra.Command{
		Use:   "list [pastes|urls]",
		Short: "List your pastes or shortened URLs",
		Args:  cobra.ExactArgs(1),
		RunE:  List,
	}

	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&limit, "per-page", 10, "Number of items per page")

	return cmd
}

func List(cmd *cobra.Command, args []string) error {
	listType := "pastes"
	if len(args) > 0 {
		listType = args[0]
	}

	page, err := cmd.Flags().GetInt("page")
	if err != nil {
		return err
	}

	perPage, err := cmd.Flags().GetInt("per-page")
	if err != nil {
		return err
	}

	switch listType {
	case "pastes":
		resp, err := client.ListPastes(page, perPage)
		if err != nil {
			return fmt.Errorf("error listing pastes: %w", err)
		}

		if !resp.Success {
			return fmt.Errorf("error listing pastes: %s", resp.Error)
		}

		fmt.Fprintln(cmd.OutOrStdout(), theme.Title.Render("Your Pastes"))
		for _, item := range resp.Data.Items {
			createdAt, err := time.Parse(time.RFC3339, item.CreatedAt)
			if err != nil {
				createdAt = time.Time{}
			}

			fmt.Fprintln(cmd.OutOrStdout(), theme.FormatKeyValue("ID", item.Id))
			fmt.Fprintln(cmd.OutOrStdout(), theme.FormatKeyValue("Filename", item.Filename))
			fmt.Fprintf(cmd.OutOrStdout(), "%s %d bytes\n", theme.ListItemKey.Render("Size:"), item.Size)
			fmt.Fprintln(cmd.OutOrStdout(), theme.FormatKeyValue("Created", createdAt.Format(time.RFC3339)))
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", theme.ListItemKey.Render("URL:"), theme.FormatURL(item.URL))
			fmt.Fprintln(cmd.OutOrStdout())
		}

	case "urls":
		resp, err := client.ListURLs(page, perPage)
		if err != nil {
			return fmt.Errorf("error listing URLs: %w", err)
		}

		if !resp.Success {
			return fmt.Errorf("error listing URLs: %s", resp.Error)
		}

		fmt.Fprintln(cmd.OutOrStdout(), theme.Title.Render("Your Shortened URLs"))
		for _, item := range resp.Data.Items {
			createdAt, err := time.Parse(time.RFC3339, item.CreatedAt)
			if err != nil {
				createdAt = time.Time{}
			}

			fmt.Fprintln(cmd.OutOrStdout(), theme.FormatKeyValue("ID", item.Id))
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", theme.ListItemKey.Render("Short URL:"), theme.FormatURL(item.ShortURL))
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", theme.ListItemKey.Render("Original URL:"), theme.FormatURL(item.OriginalURL))
			fmt.Fprintln(cmd.OutOrStdout(), theme.FormatKeyValue("Created", createdAt.Format(time.RFC3339)))
			fmt.Fprintln(cmd.OutOrStdout())
		}

	default:
		return fmt.Errorf("%s", theme.FormatError("Invalid list type. Must be 'pastes' or 'urls'"))
	}

	return nil
}

func NewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a paste or shortened URL",
		Args:  cobra.ExactArgs(1),
		RunE:  Delete,
	}

	return cmd
}

func Delete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	resp, err := client.Delete(args[0])
	if err != nil {
		return fmt.Errorf("error deleting content: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("error deleting content: %s", resp.Error)
	}

	fmt.Fprintln(cmd.OutOrStdout(), resp.Message)
	return nil
}

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	getCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			value := viper.GetString(args[0])
			if value == "" {
				fmt.Fprintln(cmd.OutOrStdout(), theme.FormatWarning("Config value not set"))
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), theme.FormatKeyValue(args[0], value))
			return nil
		},
	}

	setCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(args[0], args[1])
			if err := viper.WriteConfig(); err != nil {
				if os.IsNotExist(err) {
					configDir := filepath.Dir(viper.ConfigFileUsed())
					if err := os.MkdirAll(configDir, 0755); err != nil {
						return fmt.Errorf(theme.FormatError("Could not create config directory: %v"), err)
					}
					if err := viper.WriteConfigAs(viper.ConfigFileUsed()); err != nil {
						return fmt.Errorf(theme.FormatError("Could not write config file: %v"), err)
					}
				} else {
					return fmt.Errorf(theme.FormatError("Could not write config file: %v"), err)
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), theme.FormatSuccess("Config value '%s' set to '%s'\n"), args[0], args[1])
			return nil
		},
	}

	cmd.AddCommand(getCmd, setCmd)
	return cmd
}
