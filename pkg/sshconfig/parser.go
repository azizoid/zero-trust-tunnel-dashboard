package sshconfig

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type Config struct {
	Host         string
	HostName     string
	User         string
	IdentityFile string
	Port         int
}

func ParseSSHConfig(host string) (*Config, error) {
	configPath := os.Getenv("SSH_CONFIG")
	if configPath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("failed to get current user: %w", err)
		}
		configPath = filepath.Join(usr.HomeDir, ".ssh", "config")
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSH config: %w", err)
	}
	defer func() {
		_ = file.Close() //nolint:errcheck // Ignore error on close
	}()

	config := &Config{
		Host: host,
		Port: 22,
	}

	scanner := bufio.NewScanner(file)
	inHostBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(strings.ToLower(line), "host ") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			hostPatterns := parts[1:]
			matches := false
			for _, pattern := range hostPatterns {
				if pattern == host || pattern == "*" {
					matches = true
					break
				}
				if strings.HasPrefix(pattern, "*") && strings.HasSuffix(host, strings.TrimPrefix(pattern, "*")) {
					matches = true
					break
				}
			}

			if matches {
				inHostBlock = true
			} else {
				inHostBlock = false
			}
			continue
		}

		if inHostBlock {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			key := strings.ToLower(parts[0])
			value := strings.Join(parts[1:], " ")

			switch key {
			case "hostname":
				config.HostName = value
			case "user":
				config.User = value
			case "identityfile":
				if strings.HasPrefix(value, "~") {
					usr, err := user.Current()
					if err == nil {
						value = filepath.Join(usr.HomeDir, strings.TrimPrefix(value, "~"))
					}
				}
				config.IdentityFile = value
			case "port":
				_, _ = fmt.Sscanf(value, "%d", &config.Port) // Ignore parse error, defaults to 22
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading SSH config: %w", err)
	}

	if config.HostName == "" {
		config.HostName = host
	}

	return config, nil
}
