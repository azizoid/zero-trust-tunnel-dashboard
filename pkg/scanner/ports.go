package scanner

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/ssh"
)

type Scanner struct {
	sshClient    *ssh.Client
	server       string
	user         string
	keyPath      string
	useHostAlias bool
	hostAlias    string
	insecure     bool
}

func NewScanner(server, user, keyPath string) *Scanner {
	s := &Scanner{
		server:       server,
		user:         user,
		keyPath:      keyPath,
		useHostAlias: false,
	}
	s.initClient()
	return s
}

func NewScannerWithHost(hostAlias string) *Scanner {
	s := &Scanner{
		useHostAlias: true,
		hostAlias:    hostAlias,
	}
	s.initClient()
	return s
}

func (s *Scanner) SetInsecure(insecure bool) {
	s.insecure = insecure
	s.initClient()
}

func (s *Scanner) initClient() {
	config := ssh.Config{
		Server:       s.server,
		User:         s.user,
		KeyPath:      s.keyPath,
		UseHostAlias: s.useHostAlias,
		HostAlias:    s.hostAlias,
		Insecure:     s.insecure,
	}
	s.sshClient = ssh.NewClient(config)
}

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

func (s *Scanner) scanWithSS(portRange string) ([]int, error) {
	cmd := s.sshClient.BuildCommand("ss -tlnp")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseSSOutput(string(output), portRange)
}

func (s *Scanner) scanWithNetstat(portRange string) ([]int, error) {
	cmd := s.sshClient.BuildCommand("netstat -tlnp")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseNetstatOutput(string(output), portRange)
}

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

func parsePortRange(portRange string) (minPort, maxPort int) {
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
