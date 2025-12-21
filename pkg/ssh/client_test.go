package ssh

import (
	"context"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	config := Config{
		Server:  "example.com",
		User:    "testuser",
		KeyPath: "/path/to/key",
	}

	client := NewClient(config)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.config.Server != "example.com" {
		t.Errorf("Expected Server to be 'example.com', got '%s'", client.config.Server)
	}

	if client.config.User != "testuser" {
		t.Errorf("Expected User to be 'testuser', got '%s'", client.config.User)
	}
}

func TestBuildCommand(t *testing.T) {
	config := Config{
		Server:  "example.com",
		User:    "testuser",
		KeyPath: "/path/to/key",
	}

	client := NewClient(config)
	cmd := client.BuildCommand("ls -la")

	if cmd == nil {
		t.Fatal("BuildCommand() returned nil")
	}

	if !strings.Contains(cmd.Path, "ssh") {
		t.Errorf("Expected command path to contain 'ssh', got '%s'", cmd.Path)
	}

	args := cmd.Args
	if len(args) < 4 {
		t.Errorf("Expected at least 4 args, got %d", len(args))
	}

	hasKeyFlag := false
	hasUserHost := false
	for i, arg := range args {
		if arg == "-i" && i+1 < len(args) && args[i+1] == "/path/to/key" {
			hasKeyFlag = true
		}
		if strings.Contains(arg, "testuser@example.com") {
			hasUserHost = true
		}
	}

	if !hasKeyFlag {
		t.Error("Expected -i flag with key path in command args")
	}

	if !hasUserHost {
		t.Error("Expected user@host in command args")
	}
}

func TestBuildCommandWithHostAlias(t *testing.T) {
	config := Config{
		UseHostAlias: true,
		HostAlias:    "myserver",
	}

	client := NewClient(config)
	cmd := client.BuildCommand("ls -la")

	if cmd == nil {
		t.Fatal("BuildCommand() returned nil")
	}

	args := cmd.Args
	foundHostAlias := false
	for i, arg := range args {
		if arg == "myserver" && i < len(args)-1 && args[i+1] == "ls -la" {
			foundHostAlias = true
			break
		}
	}

	if !foundHostAlias {
		t.Error("Host alias not found in command args")
	}
}

func TestBuildTunnelCommand(t *testing.T) {
	config := Config{
		Server:  "example.com",
		User:    "testuser",
		KeyPath: "/path/to/key",
	}

	client := NewClient(config)
	ctx := context.Background()
	cmd := client.BuildTunnelCommand(ctx, 9000, 3000)

	if cmd == nil {
		t.Fatal("BuildTunnelCommand() returned nil")
	}

	if !strings.Contains(cmd.Path, "ssh") {
		t.Errorf("Expected command path to contain 'ssh', got '%s'", cmd.Path)
	}

	args := cmd.Args
	foundTunnel := false
	for _, arg := range args {
		if arg == "-L" {
			foundTunnel = true
			break
		}
	}

	if !foundTunnel {
		t.Error("Tunnel flag (-L) not found in command args")
	}

	foundLocalPort := false
	for _, arg := range args {
		if arg == "9000:localhost:3000" {
			foundLocalPort = true
			break
		}
	}

	if !foundLocalPort {
		t.Error("Local port mapping not found in command args")
	}
}

func TestBuildTunnelCommandWithHostAlias(t *testing.T) {
	config := Config{
		UseHostAlias: true,
		HostAlias:    "myserver",
	}

	client := NewClient(config)
	ctx := context.Background()
	cmd := client.BuildTunnelCommand(ctx, 9000, 3000)

	if cmd == nil {
		t.Fatal("BuildTunnelCommand() returned nil")
	}

	args := cmd.Args
	foundHostAlias := false
	for _, arg := range args {
		if arg == "myserver" {
			foundHostAlias = true
			break
		}
	}

	if !foundHostAlias {
		t.Error("Host alias not found in tunnel command args")
	}
}

