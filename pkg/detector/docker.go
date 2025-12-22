package detector

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Package-level regex compilation for performance
var (
	portRegex              = regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+:)?(\d+)->(\d+)/tcp`)
	containerOnlyPortRegex = regexp.MustCompile(`(\d+)/tcp`)
)

// DockerService represents a Docker container with port information
type DockerService struct {
	ContainerName string
	Image         string
	Port          int
	PortMapping   string
	Network       string
	HasPorts      bool
	ExposedToHost bool
}

// PortInfo represents extracted port information from Docker port mappings
type PortInfo struct {
	Ports         []int
	ExposedToHost bool
}

// buildSSHCommand constructs an SSH command for remote execution
func buildSSHCommand(server, user, keyPath string, useHostAlias bool, hostAlias, cmd string) *exec.Cmd {
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
	}

	if useHostAlias {
		args = append(args, hostAlias, cmd)
	} else {
		if keyPath != "" {
			args = append(args, "-i", keyPath)
		}
		args = append(args, fmt.Sprintf("%s@%s", user, server), cmd)
	}

	return exec.Command("ssh", args...)
}

// executeDockerPS runs docker ps remotely via SSH and returns the output
func executeDockerPS(server, user, keyPath string, useHostAlias bool, hostAlias string) (string, error) {
	cmd := buildSSHCommand(server, user, keyPath, useHostAlias, hostAlias,
		"docker ps --format '{{.Names}}|{{.Image}}|{{.Ports}}|{{.Networks}}'")

	output, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("docker ps failed: %s: %w", string(ee.Stderr), err)
		}
		return "", fmt.Errorf("failed to run docker ps: %w", err)
	}

	return string(output), nil
}

// DetectDockerServices detects Docker services with ports exposed to the host
func DetectDockerServices(server, user, keyPath string, useHostAlias bool, hostAlias string) (map[int]*DockerService, error) {
	output, err := executeDockerPS(server, user, keyPath, useHostAlias, hostAlias)
	if err != nil {
		return nil, err
	}

	return filterExposedContainers(parseDockerContainers(output)), nil
}

// GetAllDockerContainers returns all Docker containers regardless of port exposure
func GetAllDockerContainers(server, user, keyPath string, useHostAlias bool, hostAlias string) ([]*DockerService, error) {
	output, err := executeDockerPS(server, user, keyPath, useHostAlias, hostAlias)
	if err != nil {
		return nil, err
	}

	return parseDockerContainers(output), nil
}

// filterExposedContainers filters containers to only those with ports exposed to the host
func filterExposedContainers(containers []*DockerService) map[int]*DockerService {
	services := make(map[int]*DockerService)
	for _, container := range containers {
		if container.HasPorts && container.Port > 0 && container.ExposedToHost {
			services[container.Port] = container
		}
	}
	return services
}

// parseDockerContainers parses docker ps output into DockerService structs
func parseDockerContainers(output string) []*DockerService {
	var containers []*DockerService
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		containerName := strings.TrimSpace(parts[0])
		image := strings.TrimSpace(parts[1])
		portsStr := strings.TrimSpace(parts[2])
		network := ""
		if len(parts) >= 4 {
			network = strings.TrimSpace(parts[3])
		}

		portInfo := extractPorts(portsStr)

		if len(portInfo.Ports) > 0 {
			for _, port := range portInfo.Ports {
				containers = append(containers, &DockerService{
					ContainerName: containerName,
					Image:         image,
					Port:          port,
					PortMapping:   portsStr,
					Network:       network,
					HasPorts:      true,
					ExposedToHost: portInfo.ExposedToHost,
				})
			}
		} else {
			containers = append(containers, &DockerService{
				ContainerName: containerName,
				Image:         image,
				Port:          0,
				PortMapping:   "",
				Network:       network,
				HasPorts:      false,
				ExposedToHost: false,
			})
		}
	}

	return containers
}

// extractPorts extracts port information from Docker port mapping string
func extractPorts(portsStr string) PortInfo {
	var ports []int
	exposedToHost := false

	// Match port mappings like "0.0.0.0:8080->80/tcp" or "8080->80/tcp"
	matches := portRegex.FindAllStringSubmatch(portsStr, -1)
	for _, match := range matches {
		if len(match) >= 4 {
			hostPort := match[2]
			containerPort := match[3]

			// Prefer host port if available, otherwise use container port
			port, err := parsePort(hostPort)
			if err == nil && port > 0 {
				ports = append(ports, port)
				exposedToHost = true
			} else {
				port, err = parsePort(containerPort)
				if err == nil && port > 0 {
					ports = append(ports, port)
					// Container port only, not exposed to host
				}
			}
		}
	}

	// If no exposed ports found, check for container-only ports like "80/tcp"
	if len(ports) == 0 {
		containerMatches := containerOnlyPortRegex.FindAllStringSubmatch(portsStr, -1)
		for _, match := range containerMatches {
			if len(match) >= 2 {
				port, err := parsePort(match[1])
				if err == nil && port > 0 {
					ports = append(ports, port)
					exposedToHost = false
				}
			}
		}
	}

	return PortInfo{
		Ports:         ports,
		ExposedToHost: exposedToHost,
	}
}

// parsePort parses a port string to an integer
func parsePort(portStr string) (int, error) {
	portStr = strings.TrimSpace(portStr)
	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	return port, err
}

// serviceMatcher defines how to match and identify a service from Docker container info
type serviceMatcher struct {
	Match func(image, containerName string) bool
	Type  string
	Name  string
	Desc  func(image string) string
}

// normalizeImageName extracts the base image name without tag
func normalizeImageName(image string) string {
	if strings.Contains(image, ":") {
		parts := strings.Split(image, ":")
		return parts[0]
	}
	return image
}

// serviceMatchers is the registry of known service types
var serviceMatchers = []serviceMatcher{
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "grafana") || strings.Contains(nameLower, "grafana")
		},
		Type: "grafana",
		Name: "Grafana",
		Desc: func(img string) string {
			return fmt.Sprintf("Grafana Dashboard (%s)", normalizeImageName(img))
		},
	},
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "prometheus") || strings.Contains(nameLower, "prometheus")
		},
		Type: "prometheus",
		Name: "Prometheus",
		Desc: func(img string) string {
			return fmt.Sprintf("Prometheus Metrics Server (%s)", normalizeImageName(img))
		},
	},
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "nginx") || strings.Contains(nameLower, "nginx")
		},
		Type: "nginx",
		Name: "Nginx",
		Desc: func(img string) string {
			return fmt.Sprintf("Nginx Server (%s)", img)
		},
	},
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "postgres") || strings.Contains(nameLower, "postgres") || strings.Contains(nameLower, "db")
		},
		Type: "postgres",
		Name: "PostgreSQL",
		Desc: func(img string) string {
			return fmt.Sprintf("PostgreSQL Database (%s)", img)
		},
	},
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "redis") || strings.Contains(nameLower, "redis")
		},
		Type: "redis",
		Name: "Redis",
		Desc: func(img string) string {
			return fmt.Sprintf("Redis Cache (%s)", img)
		},
	},
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "mysql") || strings.Contains(nameLower, "mysql")
		},
		Type: "mysql",
		Name: "MySQL",
		Desc: func(img string) string {
			return fmt.Sprintf("MySQL Database (%s)", img)
		},
	},
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "mongodb") || strings.Contains(nameLower, "mongodb")
		},
		Type: "mongodb",
		Name: "MongoDB",
		Desc: func(img string) string {
			return fmt.Sprintf("MongoDB Database (%s)", img)
		},
	},
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "jupyter") || strings.Contains(nameLower, "jupyter")
		},
		Type: "jupyter",
		Name: "Jupyter",
		Desc: func(img string) string {
			return fmt.Sprintf("Jupyter Notebook (%s)", img)
		},
	},
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "jenkins") || strings.Contains(nameLower, "jenkins")
		},
		Type: "jenkins",
		Name: "Jenkins",
		Desc: func(img string) string {
			return fmt.Sprintf("Jenkins CI/CD (%s)", img)
		},
	},
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "elasticsearch") || strings.Contains(nameLower, "elasticsearch")
		},
		Type: "elasticsearch",
		Name: "Elasticsearch",
		Desc: func(img string) string {
			return fmt.Sprintf("Elasticsearch (%s)", img)
		},
	},
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "kibana") || strings.Contains(nameLower, "kibana")
		},
		Type: "kibana",
		Name: "Kibana",
		Desc: func(img string) string {
			return fmt.Sprintf("Kibana (%s)", img)
		},
	},
	{
		Match: func(img, name string) bool {
			imgLower := strings.ToLower(img)
			nameLower := strings.ToLower(name)
			return strings.Contains(imgLower, "rabbitmq") || strings.Contains(nameLower, "rabbitmq")
		},
		Type: "rabbitmq",
		Name: "RabbitMQ",
		Desc: func(img string) string {
			return fmt.Sprintf("RabbitMQ (%s)", img)
		},
	},
	{
		Match: func(_ string, name string) bool {
			nameLower := strings.ToLower(name)
			return strings.Contains(nameLower, "app") || strings.Contains(nameLower, "api") || strings.Contains(nameLower, "service")
		},
		Type: "application",
		Name: "", // Keep original container name
		Desc: func(image string) string {
			return fmt.Sprintf("Application Service (%s)", image)
		},
	},
}

// IdentifyServiceFromDocker identifies a service type from Docker container information
// Returns a Service with URL set to empty string (URL generation belongs in dashboard layer)
func IdentifyServiceFromDocker(dockerSvc *DockerService) *Service {
	imageName := normalizeImageName(dockerSvc.Image)
	imageLower := strings.ToLower(dockerSvc.Image)
	containerLower := strings.ToLower(dockerSvc.ContainerName)

	// Try each matcher in order
	for _, matcher := range serviceMatchers {
		if matcher.Match(imageLower, containerLower) {
			name := matcher.Name
			if name == "" {
				name = dockerSvc.ContainerName
			}
			return &Service{
				Port:        dockerSvc.Port,
				Name:        name,
				Type:        matcher.Type,
				URL:         "", // URL generation belongs in dashboard/view layer
				Description: matcher.Desc(dockerSvc.Image),
				Network:     dockerSvc.Network,
			}
		}
	}

	// Default: generic Docker container
	return &Service{
		Port:        dockerSvc.Port,
		Name:        dockerSvc.ContainerName,
		Type:        "docker",
		URL:         "", // URL generation belongs in dashboard/view layer
		Description: fmt.Sprintf("Docker container: %s (%s)", dockerSvc.ContainerName, imageName),
		Network:     dockerSvc.Network,
	}
}
