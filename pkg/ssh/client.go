package ssh

import (
	"context"
	"fmt"
	"os/exec"
)

type Config struct {
	Server       string
	User         string
	KeyPath      string
	UseHostAlias bool
	HostAlias    string
	Insecure     bool
}

type Client struct {
	config Config
}

func NewClient(config Config) *Client {
	return &Client{config: config}
}

func (c *Client) BuildCommand(remoteCmd string) *exec.Cmd {
	args := c.buildSSHArgs(remoteCmd)
	return exec.Command("ssh", args...)
}

// BuildCommandWithContext builds an SSH command with context support for cancellation.
func (c *Client) BuildCommandWithContext(ctx context.Context, remoteCmd string) *exec.Cmd {
	args := c.buildSSHArgs(remoteCmd)
	return exec.CommandContext(ctx, "ssh", args...)
}

func (c *Client) BuildTunnelCommand(ctx context.Context, localPort, remotePort int) *exec.Cmd {
	args := []string{
		"-L", fmt.Sprintf("%d:localhost:%d", localPort, remotePort),
		"-N",
	}

	if c.config.Insecure {
		args = append(args,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
		)
	}

	args = append(args, "-o", "LogLevel=ERROR")

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

func (c *Client) buildSSHArgs(remoteCmd string) []string {
	var args []string

	if c.config.Insecure {
		args = append(args,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
		)
	}

	args = append(args, "-o", "LogLevel=ERROR")

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
