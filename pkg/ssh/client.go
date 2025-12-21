package ssh

import (
	"context"
	"fmt"
	"os/exec"
)

// Config represents SSH connection configuration
type Config struct {
	Server       string
	User         string
	KeyPath      string
	UseHostAlias bool
	HostAlias    string
}

// Client handles SSH command execution
type Client struct {
	config Config
}

// NewClient creates a new SSH client
func NewClient(config Config) *Client {
	return &Client{config: config}
}

// BuildCommand builds an SSH command to execute on the remote server
func (c *Client) BuildCommand(remoteCmd string) *exec.Cmd {
	args := c.buildSSHArgs(remoteCmd)
	return exec.Command("ssh", args...)
}

// BuildCommandWithContext builds an SSH command with context for cancellation
func (c *Client) BuildCommandWithContext(ctx context.Context, remoteCmd string) *exec.Cmd {
	args := c.buildSSHArgs(remoteCmd)
	return exec.CommandContext(ctx, "ssh", args...)
}

// BuildTunnelCommand builds an SSH tunnel command (-L flag)
func (c *Client) BuildTunnelCommand(ctx context.Context, localPort, remotePort int) *exec.Cmd {
	args := []string{
		"-L", fmt.Sprintf("%d:localhost:%d", localPort, remotePort),
		"-N",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
	}

	if c.config.UseHostAlias {
		args = append(args, c.config.HostAlias)
	} else {
		if c.config.KeyPath != "" {
			args = append(args, "-i", c.config.KeyPath)
		}
		args = append(args, fmt.Sprintf("%s@%s", c.config.User, c.config.Server))
	}

	return exec.CommandContext(ctx, "ssh", args...)
}

// buildSSHArgs builds the base SSH arguments
func (c *Client) buildSSHArgs(remoteCmd string) []string {
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
	}

	if c.config.UseHostAlias {
		args = append(args, c.config.HostAlias, remoteCmd)
	} else {
		if c.config.KeyPath != "" {
			args = append(args, "-i", c.config.KeyPath)
		}
		args = append(args, fmt.Sprintf("%s@%s", c.config.User, c.config.Server), remoteCmd)
	}

	return args
}

