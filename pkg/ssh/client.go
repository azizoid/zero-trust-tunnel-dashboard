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

func (c *Client) BuildCommandWithContext(ctx context.Context, remoteCmd string) *exec.Cmd {
	args := c.buildSSHArgs(remoteCmd)
	return exec.CommandContext(ctx, "ssh", args...)
}

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
