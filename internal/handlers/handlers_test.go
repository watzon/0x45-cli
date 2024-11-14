package handlers

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestUploadCmd(t *testing.T) {
	// Set up test environment
	viper.Set("api_url", "https://0x45.st")
	viper.Set("api_key", "test-key")

	// Create a temporary test file
	content := []byte("test content")
	tmpfile := t.TempDir() + "/test.txt"
	if err := os.WriteFile(tmpfile, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Create buffer for output capture
	var out bytes.Buffer

	// Create root command
	rootCmd := &cobra.Command{Use: "test"}
	rootCmd.SetOut(&out)

	// Add upload command
	uploadCmd := NewUploadCmd()
	rootCmd.AddCommand(uploadCmd)

	// Test upload command
	rootCmd.SetArgs([]string{"upload", tmpfile, "--private", "--expires", "24h"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	// Check output
	output := out.String()
	if output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestShortenCmd(t *testing.T) {
	// Set up test environment
	viper.Set("api_url", "https://0x45.st")
	viper.Set("api_key", "test-key")

	// Create buffer for output capture
	var out bytes.Buffer

	// Create root command
	rootCmd := &cobra.Command{Use: "test"}
	rootCmd.SetOut(&out)

	// Add shorten command
	shortenCmd := NewShortenCmd()
	rootCmd.AddCommand(shortenCmd)

	// Test shorten command
	rootCmd.SetArgs([]string{"shorten", "https://example.com", "--private", "--expires", "24h"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	// Check output
	output := out.String()
	if output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestListCmd(t *testing.T) {
	// Set up test environment
	viper.Set("api_url", "https://0x45.st")
	viper.Set("api_key", "test-key")

	// Create buffer for output capture
	var out bytes.Buffer

	// Create root command
	rootCmd := &cobra.Command{Use: "test"}
	rootCmd.SetOut(&out)

	// Add list command
	listCmd := NewListCmd()
	rootCmd.AddCommand(listCmd)

	// Test list command
	rootCmd.SetArgs([]string{"list", "pastes", "--page", "1", "--limit", "10"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	// Check output
	output := out.String()
	if output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestDeleteCmd(t *testing.T) {
	// Set up test environment
	viper.Set("api_url", "https://0x45.st")
	viper.Set("api_key", "test-key")

	// Create buffer for output capture
	var out bytes.Buffer

	// Create root command
	rootCmd := &cobra.Command{Use: "test"}
	rootCmd.SetOut(&out)

	// Add delete command
	deleteCmd := NewDeleteCmd()
	rootCmd.AddCommand(deleteCmd)

	// Test delete command
	rootCmd.SetArgs([]string{"delete", "abc123"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	// Check output
	output := out.String()
	if output == "" {
		t.Error("Expected non-empty output")
	}
}
