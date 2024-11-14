package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/watzon/0x45-cli/internal/handlers"
	"github.com/watzon/0x45-cli/internal/theme"
)

var cfgFile string

func main() {
	rootCmd := &cobra.Command{
		Use:   "0x45",
		Short: theme.Title.Render("A CLI client for 0x45.st"),
		Long: theme.InfoBox.Render(`0x45 is a command line interface for 0x45.st, a file and URL sharing service.
It allows you to upload files, shorten URLs, and manage your content.`),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return validateAPIKey()
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.0x45.yaml)")

	rootCmd.AddCommand(
		handlers.NewConfigCmd(),
		handlers.NewUploadCmd(),
		handlers.NewShortenCmd(),
		handlers.NewListCmd(),
		handlers.NewDeleteCmd(),
	)

	cobra.OnInitialize(initConfig)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(theme.FormatError(err.Error()))
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".0x45")
	}

	viper.SetEnvPrefix("OX45")
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("api_url", "https://0x45.st")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println(theme.FormatError(fmt.Sprintf("Error reading config file: %v", err)))
		}
	} else {
		fmt.Println(theme.FormatSuccess(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed())))
	}
}

func validateAPIKey() error {
	if viper.GetString("api_key") == "" {
		return fmt.Errorf("%s", theme.RenderErrorBox("API key not set. Run '0x45 config set api_key YOUR_API_KEY' to set it"))
	}
	return nil
}
