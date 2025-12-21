package detector

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// DockerService represents a service discovered via Docker
type DockerService struct {
	ContainerName string
	Image         string
	Port          int
	PortMapping   string
	Network       string
	HasPorts      bool
	ExposedToHost bool // true if port is exposed to host (has -p mapping), false if only EXPOSE in Dockerfile
}

// DetectDockerServices detects services by running docker ps over SSH
func DetectDockerServices(server, user, keyPath string, useHostAlias bool, hostAlias string) (map[int]*DockerService, error) {
	var cmd *exec.Cmd
	if useHostAlias {
		args := []string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "LogLevel=ERROR",
			hostAlias,
			"docker ps --format '{{.Names}}|{{.Image}}|{{.Ports}}|{{.Networks}}'",
		}
		cmd = exec.Command("ssh", args...)
	} else {
		args := []string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "LogLevel=ERROR",
		}
		if keyPath != "" {
			args = append(args, "-i", keyPath)
		}
		args = append(args, fmt.Sprintf("%s@%s", user, server), "docker ps --format '{{.Names}}|{{.Image}}|{{.Ports}}|{{.Networks}}'")
		cmd = exec.Command("ssh", args...)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run docker ps: %w", err)
	}

	return parseDockerPS(string(output)), nil
}

// GetAllDockerContainers gets all containers with their info (including those without exposed ports)
func GetAllDockerContainers(server, user, keyPath string, useHostAlias bool, hostAlias string) ([]*DockerService, error) {
	var cmd *exec.Cmd
	if useHostAlias {
		args := []string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "LogLevel=ERROR",
			hostAlias,
			"docker ps --format '{{.Names}}|{{.Image}}|{{.Ports}}|{{.Networks}}'",
		}
		cmd = exec.Command("ssh", args...)
	} else {
		args := []string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "LogLevel=ERROR",
		}
		if keyPath != "" {
			args = append(args, "-i", keyPath)
		}
		args = append(args, fmt.Sprintf("%s@%s", user, server), "docker ps --format '{{.Names}}|{{.Image}}|{{.Ports}}|{{.Networks}}'")
		cmd = exec.Command("ssh", args...)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run docker ps: %w", err)
	}

	return parseAllDockerContainers(string(output)), nil
}

// parseDockerPS parses docker ps output and returns only containers with ports exposed to host
func parseDockerPS(output string) map[int]*DockerService {
	services := make(map[int]*DockerService)
	allContainers := parseAllDockerContainers(output)
	
	for _, container := range allContainers {
		// Only include ports that are exposed to the host (not just EXPOSE in Dockerfile)
		if container.HasPorts && container.Port > 0 && container.ExposedToHost {
			services[container.Port] = container
		}
	}
	
	return services
}

// parseAllDockerContainers parses docker ps output and returns all containers
func parseAllDockerContainers(output string) []*DockerService {
	var containers []*DockerService
	lines := strings.Split(output, "\n")

	portRegex := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+:)?(\d+)->(\d+)/tcp`)
	localhostPortRegex := regexp.MustCompile(`127\.0\.0\.1:(\d+)->(\d+)/tcp`)

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

		hasPorts := false
		var foundPorts []int

		exposedToHost := false
		matches := portRegex.FindAllStringSubmatch(portsStr, -1)
		for _, match := range matches {
			if len(match) >= 4 {
				hostPort := match[2]
				containerPort := match[3]
				
				port, err := parsePort(hostPort)
				if err == nil && port > 0 {
					foundPorts = append(foundPorts, port)
					hasPorts = true
					exposedToHost = true
				} else {
					port, err = parsePort(containerPort)
					if err == nil && port > 0 {
						foundPorts = append(foundPorts, port)
						hasPorts = true
					}
				}
			}
		}
		
		localhostMatches := localhostPortRegex.FindAllStringSubmatch(portsStr, -1)
		for _, match := range localhostMatches {
			if len(match) >= 3 {
				hostPort := match[1]
				containerPort := match[2]
				
				port, err := parsePort(hostPort)
				if err == nil && port > 0 {
					foundPorts = append(foundPorts, port)
					hasPorts = true
					exposedToHost = true
				} else {
					port, err = parsePort(containerPort)
					if err == nil && port > 0 {
						foundPorts = append(foundPorts, port)
						hasPorts = true
					}
				}
			}
		}
		
		containerOnlyPortRegex := regexp.MustCompile(`(\d+)/tcp`)
		if !hasPorts {
			containerMatches := containerOnlyPortRegex.FindAllStringSubmatch(portsStr, -1)
			for _, match := range containerMatches {
				if len(match) >= 2 {
					port, err := parsePort(match[1])
					if err == nil && port > 0 {
						foundPorts = append(foundPorts, port)
						hasPorts = true
						exposedToHost = false
					}
				}
			}
		}

		if hasPorts {
			for _, port := range foundPorts {
				containers = append(containers, &DockerService{
					ContainerName: containerName,
					Image:         image,
					Port:          port,
					PortMapping:   portsStr,
					Network:       network,
					HasPorts:      true,
					ExposedToHost: exposedToHost,
				})
			}
		}
		
		if !hasPorts {
			containers = append(containers, &DockerService{
				ContainerName: containerName,
				Image:         image,
				Port:          0,
				PortMapping:   "",
				Network:       network,
				HasPorts:      false,
			})
		}
	}

	return containers
}

func parsePort(portStr string) (int, error) {
	portStr = strings.TrimSpace(portStr)
	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	return port, err
}

// IdentifyServiceFromDocker identifies service type from Docker container info
func IdentifyServiceFromDocker(dockerSvc *DockerService) *Service {
	imageName := dockerSvc.Image
	if strings.Contains(imageName, ":") {
		parts := strings.Split(imageName, ":")
		imageName = parts[0]
		if len(parts) > 1 && parts[1] != "latest" {
			imageName = dockerSvc.Image
		}
	}

	service := &Service{
		Port:        dockerSvc.Port,
		Name:        dockerSvc.ContainerName,
		Type:        "docker",
		URL:         fmt.Sprintf("http://localhost:%d", dockerSvc.Port),
		Description: fmt.Sprintf("Docker container: %s (%s)", dockerSvc.ContainerName, imageName),
		Network:     dockerSvc.Network,
	}

	imageLower := strings.ToLower(dockerSvc.Image)
	containerLower := strings.ToLower(dockerSvc.ContainerName)

	if strings.Contains(imageLower, "grafana") || strings.Contains(containerLower, "grafana") {
		service.Name = "Grafana"
		service.Type = "grafana"
		imageName := dockerSvc.Image
		if strings.Contains(imageName, ":") {
			imageName = strings.Split(imageName, ":")[0]
		}
		service.Description = fmt.Sprintf("Grafana Dashboard (%s)", imageName)
		return service
	}

	if strings.Contains(imageLower, "prometheus") || strings.Contains(containerLower, "prometheus") {
		service.Name = "Prometheus"
		service.Type = "prometheus"
		imageName := dockerSvc.Image
		if strings.Contains(imageName, ":") {
			imageName = strings.Split(imageName, ":")[0]
		}
		service.Description = fmt.Sprintf("Prometheus Metrics Server (%s)", imageName)
		return service
	}

	if strings.Contains(imageLower, "nginx") || strings.Contains(containerLower, "nginx") {
		service.Name = "Nginx"
		service.Type = "nginx"
		service.Description = fmt.Sprintf("Nginx Server (%s)", dockerSvc.Image)
		return service
	}

	if strings.Contains(imageLower, "postgres") || strings.Contains(containerLower, "postgres") || strings.Contains(containerLower, "db") {
		service.Name = "PostgreSQL"
		service.Type = "postgres"
		service.Description = fmt.Sprintf("PostgreSQL Database (%s)", dockerSvc.Image)
		return service
	}

	if strings.Contains(imageLower, "redis") || strings.Contains(containerLower, "redis") {
		service.Name = "Redis"
		service.Type = "redis"
		service.Description = fmt.Sprintf("Redis Cache (%s)", dockerSvc.Image)
		return service
	}

	if strings.Contains(imageLower, "mysql") || strings.Contains(containerLower, "mysql") {
		service.Name = "MySQL"
		service.Type = "mysql"
		service.Description = fmt.Sprintf("MySQL Database (%s)", dockerSvc.Image)
		return service
	}

	if strings.Contains(imageLower, "mongodb") || strings.Contains(containerLower, "mongodb") {
		service.Name = "MongoDB"
		service.Type = "mongodb"
		service.Description = fmt.Sprintf("MongoDB Database (%s)", dockerSvc.Image)
		return service
	}

	if strings.Contains(imageLower, "jupyter") || strings.Contains(containerLower, "jupyter") {
		service.Name = "Jupyter"
		service.Type = "jupyter"
		service.Description = fmt.Sprintf("Jupyter Notebook (%s)", dockerSvc.Image)
		return service
	}

	if strings.Contains(imageLower, "jenkins") || strings.Contains(containerLower, "jenkins") {
		service.Name = "Jenkins"
		service.Type = "jenkins"
		service.Description = fmt.Sprintf("Jenkins CI/CD (%s)", dockerSvc.Image)
		return service
	}

	if strings.Contains(imageLower, "elasticsearch") || strings.Contains(containerLower, "elasticsearch") {
		service.Name = "Elasticsearch"
		service.Type = "elasticsearch"
		service.Description = fmt.Sprintf("Elasticsearch (%s)", dockerSvc.Image)
		return service
	}

	if strings.Contains(imageLower, "kibana") || strings.Contains(containerLower, "kibana") {
		service.Name = "Kibana"
		service.Type = "kibana"
		service.Description = fmt.Sprintf("Kibana (%s)", dockerSvc.Image)
		return service
	}

	if strings.Contains(imageLower, "rabbitmq") || strings.Contains(containerLower, "rabbitmq") {
		service.Name = "RabbitMQ"
		service.Type = "rabbitmq"
		service.Description = fmt.Sprintf("RabbitMQ (%s)", dockerSvc.Image)
		return service
	}

	if strings.Contains(containerLower, "app") || strings.Contains(containerLower, "api") || strings.Contains(containerLower, "service") {
		service.Type = "application"
		service.Description = fmt.Sprintf("Application Service (%s)", dockerSvc.Image)
		return service
	}

	return service
}

