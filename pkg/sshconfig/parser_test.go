package sshconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSSHConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	configContent := `Host test-server
    HostName example.com
    User testuser
    IdentityFile ~/.ssh/id_test
    Port 2222

Host another-server
    HostName another.com
    User anotheruser
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	originalEnv := os.Getenv("SSH_CONFIG")
	defer os.Setenv("SSH_CONFIG", originalEnv)

	os.Setenv("SSH_CONFIG", configPath)

	config, err := ParseSSHConfig("test-server")
	if err != nil {
		t.Fatalf("ParseSSHConfig() error = %v", err)
	}

	if config.Host != "test-server" {
		t.Errorf("Expected Host to be 'test-server', got '%s'", config.Host)
	}

	if config.HostName != "example.com" {
		t.Errorf("Expected HostName to be 'example.com', got '%s'", config.HostName)
	}

	if config.User != "testuser" {
		t.Errorf("Expected User to be 'testuser', got '%s'", config.User)
	}

	if config.Port != 2222 {
		t.Errorf("Expected Port to be 2222, got %d", config.Port)
	}
}

func TestParseSSHConfig_WithComments(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	configContent := `# This is a comment
Host test-server
    # Another comment
    HostName example.com
    User testuser
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	originalEnv := os.Getenv("SSH_CONFIG")
	defer os.Setenv("SSH_CONFIG", originalEnv)

	os.Setenv("SSH_CONFIG", configPath)

	config, err := ParseSSHConfig("test-server")
	if err != nil {
		t.Fatalf("ParseSSHConfig() error = %v", err)
	}

	if config.HostName != "example.com" {
		t.Errorf("Expected HostName to be 'example.com', got '%s'", config.HostName)
	}
}

func TestParseSSHConfig_DefaultPort(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	configContent := `Host test-server
    HostName example.com
    User testuser
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	originalEnv := os.Getenv("SSH_CONFIG")
	defer os.Setenv("SSH_CONFIG", originalEnv)

	os.Setenv("SSH_CONFIG", configPath)

	config, err := ParseSSHConfig("test-server")
	if err != nil {
		t.Fatalf("ParseSSHConfig() error = %v", err)
	}

	if config.Port != 22 {
		t.Errorf("Expected default Port to be 22, got %d", config.Port)
	}
}

func TestParseSSHConfig_NonExistentHost(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	configContent := `Host other-server
    HostName example.com
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	originalEnv := os.Getenv("SSH_CONFIG")
	defer os.Setenv("SSH_CONFIG", originalEnv)

	os.Setenv("SSH_CONFIG", configPath)

	config, err := ParseSSHConfig("non-existent")
	if err != nil {
		t.Fatalf("ParseSSHConfig() error = %v", err)
	}

	if config.HostName == "" {
		t.Error("Expected HostName to default to host name when not found")
	}

	if config.HostName != "non-existent" {
		t.Errorf("Expected HostName to be 'non-existent', got '%s'", config.HostName)
	}
}

