package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/watzon/0x45-cli/internal/handlers"
)

var cfgFile string

func main() {
	rootCmd := &cobra.Command{
		Use:   "0x45",
		Short: "A CLI client for 0x45.st",
		Long: `0x45 is a command line interface for 0x45.st, a file and URL sharing service.
It allows you to upload files, shorten URLs, and manage your content.`,
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
		fmt.Println(err)
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

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func validateAPIKey() error {
	apiKey := viper.GetString("api_key")
	if apiKey == "" {
		return fmt.Errorf("API key required. Set it with: 0x45 config set api_key <your-key>")
	}
	return nil
}
