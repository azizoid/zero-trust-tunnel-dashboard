package sshconfig

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkParseSSHConfig(b *testing.B) {
	configContent := `Host example-server
    HostName example.com
    User admin
    IdentityFile ~/.ssh/id_rsa
    Port 22

Host another-server
    HostName another.com
    User user
    IdentityFile ~/.ssh/id_ed25519
    Port 2222
`

	// Create a temporary config file for testing
	// Note: This is a simplified benchmark that tests parsing logic
	// In real usage, this would read from an actual file
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate parsing by processing the config content
		lines := strings.Split(configContent, "\n")
		inHostBlock := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(strings.ToLower(line), "host ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					hostPatterns := parts[1:]
					matches := false
					for _, pattern := range hostPatterns {
						if pattern == "example-server" || pattern == "*" {
							matches = true
							break
						}
					}
					inHostBlock = matches
				}
				continue
			}
			if inHostBlock {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					key := strings.ToLower(parts[0])
					_ = key
					_ = strings.Join(parts[1:], " ")
				}
			}
		}
	}
}

func BenchmarkParseSSHConfigLarge(b *testing.B) {
	// Simulate a large SSH config with many hosts
	var lines []string
	lines = append(lines, "Host example-server")
	lines = append(lines, "    HostName example.com")
	lines = append(lines, "    User admin")
	
	for i := 0; i < 100; i++ {
		lines = append(lines, fmt.Sprintf("Host server-%d", i))
		lines = append(lines, fmt.Sprintf("    HostName server%d.example.com", i))
		lines = append(lines, fmt.Sprintf("    User user%d", i))
		lines = append(lines, fmt.Sprintf("    Port %d", 2200+i))
	}
	
	configContent := strings.Join(lines, "\n")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lines := strings.Split(configContent, "\n")
		inHostBlock := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(strings.ToLower(line), "host ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					hostPatterns := parts[1:]
					matches := false
					for _, pattern := range hostPatterns {
						if pattern == "example-server" || pattern == "*" {
							matches = true
							break
						}
					}
					inHostBlock = matches
				}
				continue
			}
			if inHostBlock {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					key := strings.ToLower(parts[0])
					_ = key
					_ = strings.Join(parts[1:], " ")
				}
			}
		}
	}
}

