package handlers

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/watzon/0x45-cli/internal/client"
)

var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00ff00")).
		Bold(true)
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
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "File uploaded successfully!\n")
			fmt.Fprintf(cmd.OutOrStdout(), "URL: %s\n", resp.URL)
			fmt.Fprintf(cmd.OutOrStdout(), "Delete URL: %s\n", resp.DeleteURL)
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
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "URL shortened successfully!\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Short URL: %s\n", resp.URL)
			fmt.Fprintf(cmd.OutOrStdout(), "Delete URL: %s\n", resp.DeleteURL)
			return nil
		},
	}

	cmd.Flags().BoolVar(&private, "private", false, "Make the shortened URL private")
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
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Your Pastes:\n")
				for _, item := range resp.Data.Items {
					fmt.Fprintf(cmd.OutOrStdout(), "- %s (%s) - %s\n", item.Filename, item.CreatedAt, item.URL)
				}

			case "urls":
				resp, err := client.ListURLs(page, limit)
				if err != nil {
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Your Shortened URLs:\n")
				for _, item := range resp.Data.Items {
					fmt.Fprintf(cmd.OutOrStdout(), "- %s -> %s\n", item.ShortURL, item.URL)
				}

			default:
				return fmt.Errorf("invalid list type: %s", args[0])
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
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted content: %s\n", resp.Message)
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

	cmd.AddCommand(
		&cobra.Command{
			Use:   "set [key] [value]",
			Short: "Set a config value",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				viper.Set(args[0], args[1])
				if err := viper.WriteConfig(); err != nil {
					if os.IsNotExist(err) {
						if err := viper.SafeWriteConfig(); err != nil {
							return err
						}
					} else {
						return err
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Config value '%s' set to '%s'\n", args[0], args[1])
				return nil
			},
		},
		&cobra.Command{
			Use:   "get [key]",
			Short: "Get a config value",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if !viper.IsSet(args[0]) {
					return fmt.Errorf("config key '%s' not found", args[0])
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%v\n", viper.Get(args[0]))
				return nil
			},
		},
	)

	return cmd
}
