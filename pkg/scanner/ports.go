package scanner

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"zero-trust-dashboard/pkg/ssh"
)

// Scanner scans for open ports on a remote server via SSH
type Scanner struct {
	sshClient *ssh.Client
}

// NewScanner creates a new port scanner
func NewScanner(server, user, keyPath string) *Scanner {
	config := ssh.Config{
		Server:       server,
		User:         user,
		KeyPath:      keyPath,
		UseHostAlias: false,
	}
	return &Scanner{
		sshClient: ssh.NewClient(config),
	}
}

// NewScannerWithHost creates a new port scanner using SSH config host alias
func NewScannerWithHost(hostAlias string) *Scanner {
	config := ssh.Config{
		UseHostAlias: true,
		HostAlias:    hostAlias,
	}
	return &Scanner{
		sshClient: ssh.NewClient(config),
	}
}

// ScanPorts scans for open ports on the remote server
func (s *Scanner) ScanPorts(portRange string) ([]int, error) {
	ports, err := s.scanWithSS(portRange)
	if err != nil {
		ports, err = s.scanWithNetstat(portRange)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ports: %w", err)
		}
	}

	return ports, nil
}

// scanWithSS uses 'ss' command to find listening ports
func (s *Scanner) scanWithSS(portRange string) ([]int, error) {
	cmd := s.sshClient.BuildCommand("ss -tlnp")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseSSOutput(string(output), portRange)
}

// scanWithNetstat uses 'netstat' command to find listening ports
func (s *Scanner) scanWithNetstat(portRange string) ([]int, error) {
	cmd := s.sshClient.BuildCommand("netstat -tlnp")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseNetstatOutput(string(output), portRange)
}

// parseSSOutput parses 'ss' command output
func parseSSOutput(output, portRange string) ([]int, error) {
	var ports []int
	lines := strings.Split(output, "\n")

	portRegex := regexp.MustCompile(`:(\d+)\s`)
	minPort, maxPort := parsePortRange(portRange)

	for _, line := range lines {
		if !strings.Contains(line, "LISTEN") {
			continue
		}

		matches := portRegex.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}

		port, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		if port >= minPort && port <= maxPort {
			ports = append(ports, port)
		}
	}

	return deduplicatePorts(ports), nil
}

// parseNetstatOutput parses 'netstat' command output
func parseNetstatOutput(output, portRange string) ([]int, error) {
	var ports []int
	lines := strings.Split(output, "\n")

	portRegex := regexp.MustCompile(`:(\d+)\s`)
	minPort, maxPort := parsePortRange(portRange)

	for _, line := range lines {
		if !strings.Contains(line, "LISTEN") {
			continue
		}

		matches := portRegex.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}

		port, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		if port >= minPort && port <= maxPort {
			ports = append(ports, port)
		}
	}

	return deduplicatePorts(ports), nil
}

// parsePortRange parses port range string like "3000-9000" or "3000"
func parsePortRange(portRange string) (int, int) {
	if portRange == "" {
		return 1, 65535
	}

	parts := strings.Split(portRange, "-")
	if len(parts) == 1 {
		port, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return 1, 65535
		}
		return port, port
	}

	minPort, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	maxPort, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

	if err1 != nil || err2 != nil {
		return 1, 65535
	}

	return minPort, maxPort
}

// deduplicatePorts removes duplicate ports from slice
func deduplicatePorts(ports []int) []int {
	seen := make(map[int]bool)
	var result []int

	for _, port := range ports {
		if !seen[port] {
			seen[port] = true
			result = append(result, port)
		}
	}

	return result
}

