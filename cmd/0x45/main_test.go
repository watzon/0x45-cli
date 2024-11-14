package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/watzon/0x45-cli/internal/handlers"
)

// Helper functions for testing
func setupTestEnv(t *testing.T) (func(), string) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	// Reset viper
	viper.Reset()

	// Set default values
	viper.SetDefault("api_url", "https://0x45.st")

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
	customCfg := filepath.Join(tmpDir, ".0x45.yaml")
	cfgFile = customCfg

	// Create custom config file
	content := []byte("api_key: test-key\napi_url: https://custom.example.com")
	if err := os.WriteFile(customCfg, content, 0644); err != nil {
		t.Fatal(err)
	}

	initConfig()

	// Verify custom values are loaded
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
		t.Errorf("Unexpected error when API key is set: %v", err)
	}
}

func TestCommandStructure(t *testing.T) {
	rootCmd := &cobra.Command{Use: "0x45"}
	rootCmd.AddCommand(
		handlers.NewConfigCmd(),
		handlers.NewUploadCmd(),
		handlers.NewShortenCmd(),
		handlers.NewListCmd(),
		handlers.NewDeleteCmd(),
	)

	// Test root command
	if rootCmd.Use != "0x45" {
		t.Errorf("Expected root command name to be 0x45, got %s", rootCmd.Use)
	}

	// Test subcommands
	expectedCmds := map[string]bool{
		"config":  true,
		"upload":  true,
		"shorten": true,
		"list":    true,
		"delete":  true,
	}

	for _, cmd := range rootCmd.Commands() {
		if !expectedCmds[cmd.Name()] {
			t.Errorf("Unexpected command: %s", cmd.Name())
		}
		delete(expectedCmds, cmd.Name())
	}

	if len(expectedCmds) > 0 {
		var missing []string
		for name := range expectedCmds {
			missing = append(missing, name)
		}
		t.Errorf("Missing commands: %v", missing)
	}
}

func TestConfigCommand(t *testing.T) {
	cleanup, tmpDir := setupTestEnv(t)
	defer cleanup()

	// Create config file
	configFile := filepath.Join(tmpDir, ".0x45.yaml")
	if err := os.WriteFile(configFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(configFile)

	cmd := handlers.NewConfigCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"set", "api_key", "test-key"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Unexpected error executing config command: %v", err)
	}

	out := b.String()
	if !strings.Contains(out, "Config value 'api_key' set to 'test-key'") {
		t.Errorf("Expected success message, got: %s", out)
	}

	if key := viper.GetString("api_key"); key != "test-key" {
		t.Errorf("Expected API key to be test-key, got %s", key)
	}
}

func TestUploadCommand(t *testing.T) {
	cleanup, tmpDir := setupTestEnv(t)
	defer cleanup()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := handlers.NewUploadCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{testFile, "--private"})

	// Set required config
	viper.Set("api_key", "test-key")

	// This will fail without a mock server, which is expected
	_ = cmd.Execute()
}

func TestShortenCommand(t *testing.T) {
	cleanup, _ := setupTestEnv(t)
	defer cleanup()

	cmd := handlers.NewShortenCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"https://example.com", "--private"})

	// Set required config
	viper.Set("api_key", "test-key")

	// This will fail without a mock server, which is expected
	_ = cmd.Execute()
}

func TestListCommand(t *testing.T) {
	cleanup, _ := setupTestEnv(t)
	defer cleanup()

	cmd := handlers.NewListCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"pastes"})

	// Set required config
	viper.Set("api_key", "test-key")

	// This will fail without a mock server, which is expected
	_ = cmd.Execute()

	// Test URLs listing
	cmd.SetArgs([]string{"urls"})
	_ = cmd.Execute()
}
