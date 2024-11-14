package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Helper functions for testing
func setupTestEnv(t *testing.T) (func(), string) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	// Reset viper
	viper.Reset()

	// Return cleanup function
	return func() {
		os.Setenv("HOME", oldHome)
		viper.Reset()
	}, tmpDir
}

func TestInitConfig(t *testing.T) {
	cleanup, tmpDir := setupTestEnv(t)
	defer cleanup()

	// Test with default config path
	cfgFile = ""
	initConfig()

	// Verify default API URL is set
	if url := viper.GetString("api_url"); url != "https://0x45.st" {
		t.Errorf("Expected default API URL to be https://0x45.st, got %s", url)
	}

	// Test with custom config file
	customCfg := filepath.Join(tmpDir, "custom.yaml")
	cfgFile = customCfg

	// Create custom config file
	content := []byte("api_key: test-key\napi_url: https://custom.example.com")
	if err := os.WriteFile(customCfg, content, 0644); err != nil {
		t.Fatal(err)
	}

	initConfig()

	if key := viper.GetString("api_key"); key != "test-key" {
		t.Errorf("Expected API key to be test-key, got %s", key)
	}
	if url := viper.GetString("api_url"); url != "https://custom.example.com" {
		t.Errorf("Expected API URL to be https://custom.example.com, got %s", url)
	}
}

func TestValidateAPIKey(t *testing.T) {
	cleanup, _ := setupTestEnv(t)
	defer cleanup()

	// Test without API key
	viper.Set("api_key", "")
	if err := validateAPIKey(); err == nil {
		t.Error("Expected error when API key is not set")
	}

	// Test with API key
	viper.Set("api_key", "test-key")
	if err := validateAPIKey(); err != nil {
		t.Errorf("Expected no error with API key set, got %v", err)
	}
}

func TestCommandStructure(t *testing.T) {
	cleanup, _ := setupTestEnv(t)
	defer cleanup()

	var rootCmd = &cobra.Command{
		Use:   "0x45",
		Short: titleStyle.Render("A CLI tool for interacting with 0x45.st paste service"),
	}

	// Add all subcommands
	rootCmd.AddCommand(
		newConfigCommand(),
		newListCommand(),
		newUploadCommand(),
		newShortenCommand(),
		newDeleteCommand(),
		newKeyCommand(),
	)

	// Test config command
	if cmd, _, err := rootCmd.Find([]string{"config"}); err != nil || cmd.Name() != "config" {
		t.Error("Config command not found")
	}

	// Test list command
	if cmd, _, err := rootCmd.Find([]string{"list"}); err != nil || cmd.Name() != "list" {
		t.Error("List command not found")
	}

	// Test upload command
	if cmd, _, err := rootCmd.Find([]string{"upload"}); err != nil || cmd.Name() != "upload" {
		t.Error("Upload command not found")
	}

	// Test shorten command
	if cmd, _, err := rootCmd.Find([]string{"shorten"}); err != nil || cmd.Name() != "shorten" {
		t.Error("Shorten command not found")
	}

	// Test delete command
	if cmd, _, err := rootCmd.Find([]string{"delete"}); err != nil || cmd.Name() != "delete" {
		t.Error("Delete command not found")
	}

	// Test key command
	if cmd, _, err := rootCmd.Find([]string{"key"}); err != nil || cmd.Name() != "key" {
		t.Error("Key command not found")
	}
}

func TestConfigCommand(t *testing.T) {
	cleanup, tmpDir := setupTestEnv(t)
	defer cleanup()

	// Create config file and directory
	cfgDir := filepath.Join(tmpDir, ".config", "0x45")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfgFile = filepath.Join(cfgDir, ".0x45.yaml")

	// Create empty config file
	if err := os.WriteFile(cfgFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize config
	initConfig()

	cmd := newConfigCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Test config get
	args := []string{"get", "api_url"}
	cmd.SetArgs(args)
	viper.Set("api_url", "https://test.example.com")
	if err := cmd.Execute(); err != nil {
		t.Errorf("Failed to execute config get: %v", err)
	}
	if !strings.Contains(buf.String(), "https://test.example.com") {
		t.Errorf("Expected output to contain URL, got %s", buf.String())
	}

	// Test config set
	buf.Reset()
	args = []string{"set", "test_key", "test_value"}
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Errorf("Failed to execute config set: %v", err)
	}
	if value := viper.GetString("test_key"); value != "test_value" {
		t.Errorf("Expected test_key to be test_value, got %s", value)
	}
}

func TestUploadCommand(t *testing.T) {
	cleanup, _ := setupTestEnv(t)
	defer cleanup()

	cmd := newUploadCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Create a test file
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("test content"); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Set required config
	viper.Set("api_key", "test-key")
	viper.Set("api_url", "https://0x45.st")

	// Test upload command
	args := []string{tmpFile.Name()}
	cmd.SetArgs(args)

	// This should fail without a mock server, which is fine
	// We just want to ensure the command is structured correctly
	if err := cmd.Execute(); err != nil {
		// Expected error without mock server
		if !strings.Contains(err.Error(), "401 Unauthorized") {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestShortenCommand(t *testing.T) {
	cleanup, _ := setupTestEnv(t)
	defer cleanup()

	cmd := newShortenCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Set required config
	viper.Set("api_key", "test-key")
	viper.Set("api_url", "https://0x45.st")

	// Test shorten command
	args := []string{"https://example.com"}
	cmd.SetArgs(args)

	// This should fail without a mock server, which is fine
	// We just want to ensure the command is structured correctly
	if err := cmd.Execute(); err != nil {
		// Expected error without mock server
		if !strings.Contains(err.Error(), "401 Unauthorized") {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestListCommand(t *testing.T) {
	cleanup, _ := setupTestEnv(t)
	defer cleanup()

	cmd := newListCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Set required config
	viper.Set("api_key", "test-key")
	viper.Set("api_url", "https://0x45.st")

	// Test list urls command
	args := []string{"urls"}
	cmd.SetArgs(args)

	// This should fail without a mock server, which is fine
	// We just want to ensure the command is structured correctly
	if err := cmd.Execute(); err != nil {
		// Expected error without mock server
		if !strings.Contains(err.Error(), "401 Unauthorized") {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	// Test list pastes command
	args = []string{"pastes"}
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		// Expected error without mock server
		if !strings.Contains(err.Error(), "401 Unauthorized") {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}
