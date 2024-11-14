package handlers

import (
	"fmt"
	"os"
	"path/filepath"

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
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := client.UploadFile(args[0], private, expires)
			if err != nil {
				return fmt.Errorf(theme.FormatError("Error uploading file: %v"), err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), theme.FormatSuccess("File uploaded successfully!"))
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", theme.ListItemKey.Render("URL:"), theme.FormatURL(resp.URL))
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", theme.ListItemKey.Render("Delete URL:"), theme.FormatDeleteURL(resp.DeleteURL))
			return nil
		},
	}

	cmd.Flags().BoolVar(&private, "private", false, "Make the upload private")
	cmd.Flags().StringVar(&expires, "expires", "", "Set expiration time (e.g. 24h)")

	return cmd
}

func NewShortenCmd() *cobra.Command {
	var private bool
	var expires string

	cmd := &cobra.Command{
		Use:   "shorten [url]",
		Short: "Shorten a URL using 0x45.st",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := client.ShortenURL(args[0], private, expires)
			if err != nil {
				return fmt.Errorf(theme.FormatError("Error shortening URL: %v"), err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), theme.FormatSuccess("URL shortened successfully!"))
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", theme.ListItemKey.Render("URL:"), theme.FormatURL(resp.URL))
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", theme.ListItemKey.Render("Delete URL:"), theme.FormatDeleteURL(resp.DeleteURL))
			return nil
		},
	}

	cmd.Flags().BoolVar(&private, "private", false, "Make the URL private")
	cmd.Flags().StringVar(&expires, "expires", "", "Set expiration time (e.g. 24h)")

	return cmd
}

func NewListCmd() *cobra.Command {
	var page int
	var limit int

	cmd := &cobra.Command{
		Use:   "list [pastes|urls]",
		Short: "List your pastes or shortened URLs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "pastes":
				resp, err := client.ListPastes(page, limit)
				if err != nil {
					return fmt.Errorf(theme.FormatError("Error listing pastes: %v"), err)
				}

				if len(resp.Data.Items) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), theme.FormatWarning("No pastes found"))
					return nil
				}

				fmt.Fprintln(cmd.OutOrStdout(), theme.Title.Render("Your Pastes"))
				for _, item := range resp.Data.Items {
					fmt.Fprintln(cmd.OutOrStdout(), theme.FormatKeyValue("ID", item.Id))
					fmt.Fprintln(cmd.OutOrStdout(), theme.FormatKeyValue("Filename", item.Filename))
					fmt.Fprintf(cmd.OutOrStdout(), "%s %d bytes\n", theme.ListItemKey.Render("Size:"), item.Size)
					fmt.Fprintln(cmd.OutOrStdout(), theme.FormatKeyValue("Created", item.CreatedAt))
					fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", theme.ListItemKey.Render("URL:"), theme.FormatURL(item.URL))
					fmt.Fprintln(cmd.OutOrStdout())
				}

			case "urls":
				resp, err := client.ListURLs(page, limit)
				if err != nil {
					return fmt.Errorf(theme.FormatError("Error listing URLs: %v"), err)
				}

				if len(resp.Data.Items) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), theme.FormatWarning("No URLs found"))
					return nil
				}

				fmt.Fprintln(cmd.OutOrStdout(), theme.Title.Render("Your Shortened URLs"))
				for _, item := range resp.Data.Items {
					fmt.Fprintln(cmd.OutOrStdout(), theme.FormatKeyValue("ID", item.Id))
					fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", theme.ListItemKey.Render("Short URL:"), theme.FormatURL(item.ShortURL))
					fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", theme.ListItemKey.Render("Original URL:"), theme.FormatURL(item.OriginalURL))
					fmt.Fprintln(cmd.OutOrStdout(), theme.FormatKeyValue("Created", item.CreatedAt))
					fmt.Fprintln(cmd.OutOrStdout())
				}

			default:
				return fmt.Errorf("%s", theme.FormatError("Invalid list type. Must be 'pastes' or 'urls'"))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&limit, "limit", 10, "Number of items per page")

	return cmd
}

func NewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a paste or shortened URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := client.Delete(args[0])
			if err != nil {
				return fmt.Errorf(theme.FormatError("Error deleting content: %v"), err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), theme.FormatSuccess(resp.Message))
			return nil
		},
	}

	return cmd
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
