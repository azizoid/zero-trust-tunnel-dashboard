package detector

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Service represents a detected service
type Service struct {
	Port        int
	Name        string
	Type        string
	URL         string
	Description string
	Domain      string // Domain name if exposed via Nginx Proxy Manager
	Network     string // Docker network name
}

// Detector detects services running on ports
type Detector struct {
	timeout time.Duration
}

// NewDetector creates a new service detector
func NewDetector(timeout time.Duration) *Detector {
	if timeout == 0 {
		timeout = 3 * time.Second
	}
	return &Detector{timeout: timeout}
}

// DetectServices detects services on the given ports via localhost (tunnel)
// Uses both Docker info (if available) and HTTP probing
func (d *Detector) DetectServices(ports []int, dockerServices map[int]*DockerService) []Service {
	var services []Service
	client := &http.Client{
		Timeout: d.timeout,
	}

	for _, port := range ports {
		var service *Service

		if dockerSvc, exists := dockerServices[port]; exists {
			service = IdentifyServiceFromDocker(dockerSvc)
			httpService := d.probePort(client, port)
			if httpService != nil && httpService.Type != "unknown" {
				service.Name = httpService.Name
				service.Type = httpService.Type
				service.Description = httpService.Description
			}
		} else {
			service = d.probePort(client, port)
		}

		if service != nil {
			services = append(services, *service)
		}
	}

	return services
}

// DetectServicesFromDocker detects services only from Docker info (no HTTP probing)
func (d *Detector) DetectServicesFromDocker(ports []int, dockerServices map[int]*DockerService) []Service {
	var services []Service

	for _, port := range ports {
		if dockerSvc, exists := dockerServices[port]; exists {
			service := IdentifyServiceFromDocker(dockerSvc)
			if service != nil {
				services = append(services, *service)
			}
		} else {
			service := &Service{
				Port:        port,
				Name:        fmt.Sprintf("Service on port %d", port),
				Type:        "unknown",
				URL:         fmt.Sprintf("http://localhost:%d", port),
				Description: "No Docker container found for this port",
			}
			services = append(services, *service)
		}
	}

	return services
}

// DetectAllDockerContainers detects all Docker containers including those without exposed ports
func (d *Detector) DetectAllDockerContainers(allContainers []*DockerService) []Service {
	var services []Service

	for _, container := range allContainers {
		service := IdentifyServiceFromDocker(container)
		if service != nil {
			if !container.HasPorts || container.Port == 0 {
				service.Port = 0
				service.URL = ""
				service.Description = fmt.Sprintf("%s (No exposed ports - internal network only)", service.Description)
			}
			services = append(services, *service)
		}
	}

	return services
}

// probePort probes a port to identify the service
func (d *Detector) probePort(client *http.Client, port int) *Service {
	service := d.tryHTTP(client, port, false)
	if service != nil && service.Type != "unknown" && service.Type != "http" {
		return service
	}

	service = d.tryHTTP(client, port, true)
	if service != nil && service.Type != "unknown" && service.Type != "http" {
		return service
	}

	if service != nil && (service.Type == "http" || service.Type == "web" || service.Type == "api") {
		return service
	}

	likelyService := d.guessServiceByPort(port)
	if likelyService != nil {
		return likelyService
	}

	if service != nil {
		return service
	}

	return &Service{
		Port:        port,
		Name:        fmt.Sprintf("Service on port %d", port),
		Type:        "unknown",
		URL:         fmt.Sprintf("http://localhost:%d", port),
		Description: "Unknown service - may require manual inspection",
	}
}

// tryHTTP tries to connect via HTTP or HTTPS
func (d *Detector) tryHTTP(client *http.Client, port int, useHTTPS bool) *Service {
	protocol := "http"
	if useHTTPS {
		protocol = "https"
	}

	url := fmt.Sprintf("%s://localhost:%d", protocol, port)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}

	resp, err := client.Do(req)
	if err != nil {
		service := d.tryHTTPEndpoints(client, port, useHTTPS)
		if service != nil {
			return service
		}
		return nil
	}
	defer resp.Body.Close()

	service := d.identifyServiceFromResponse(resp, port, protocol)
	if service != nil {
		return service
	}

	return nil
}

// tryHTTPEndpoints tries common HTTP endpoints to identify the service
func (d *Detector) tryHTTPEndpoints(client *http.Client, port int, useHTTPS bool) *Service {
	protocol := "http"
	if useHTTPS {
		protocol = "https"
	}

	commonPaths := []string{"/", "/login", "/api/health", "/api", "/api/v1", "/health", "/status", "/metrics", "/graphql"}
	
	for _, path := range commonPaths {
		url := fmt.Sprintf("%s://localhost:%d%s", protocol, port, path)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode < 500 {
			service := d.identifyServiceFromResponse(resp, port, protocol)
			if service != nil && service.Type != "unknown" {
				return service
			}
		}
	}

	return nil
}

// identifyServiceFromResponse identifies service from HTTP response
func (d *Detector) identifyServiceFromResponse(resp *http.Response, port int, protocol string) *Service {
	service := &Service{
		Port: port,
		URL:  fmt.Sprintf("%s://localhost:%d", protocol, port),
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	bodyStr := strings.ToLower(string(body))

	grafanaVersion := resp.Header.Get("X-Grafana-Version")
	if grafanaVersion != "" {
		service.Name = "Grafana"
		service.Type = "grafana"
		service.Description = fmt.Sprintf("Grafana Dashboard (Version: %s)", grafanaVersion)
		return service
	}

	if strings.Contains(bodyStr, "grafana") || 
		strings.Contains(bodyStr, "grafana-app") ||
		strings.Contains(bodyStr, "login") && strings.Contains(bodyStr, "grafana") ||
		resp.Header.Get("Set-Cookie") != "" && strings.Contains(strings.ToLower(resp.Header.Get("Set-Cookie")), "grafana") {
		service.Name = "Grafana"
		service.Type = "grafana"
		service.Description = "Grafana Dashboard"
		return service
	}

	if strings.Contains(resp.Request.URL.Path, "/metrics") || strings.Contains(bodyStr, "# help") && strings.Contains(bodyStr, "# type") {
		service.Name = "Prometheus"
		service.Type = "prometheus"
		service.Description = "Prometheus Metrics Server"
		return service
	}

	if strings.Contains(bodyStr, "kubernetes") || strings.Contains(bodyStr, "k8s") {
		service.Name = "Kubernetes Dashboard"
		service.Type = "kubernetes"
		service.Description = "Kubernetes Web Dashboard"
		return service
	}

	if strings.Contains(bodyStr, "jenkins") || resp.Header.Get("X-Jenkins") != "" {
		service.Name = "Jenkins"
		service.Type = "jenkins"
		service.Description = "Jenkins CI/CD Server"
		return service
	}

	if strings.Contains(bodyStr, "jupyter") || strings.Contains(bodyStr, "notebook") {
		service.Name = "Jupyter"
		service.Type = "jupyter"
		service.Description = "Jupyter Notebook Server"
		return service
	}

	if strings.Contains(bodyStr, "react") || strings.Contains(bodyStr, "vue") || strings.Contains(bodyStr, "angular") {
		service.Name = fmt.Sprintf("Web Application (Port %d)", port)
		service.Type = "webapp"
		service.Description = "Single Page Application"
		return service
	}

	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		service.Name = fmt.Sprintf("API Service (Port %d)", port)
		service.Type = "api"
		service.Description = fmt.Sprintf("REST API Service (Status: %d)", resp.StatusCode)
		return service
	}

	if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		service.Name = fmt.Sprintf("Web Service (Port %d)", port)
		service.Type = "web"
		service.Description = fmt.Sprintf("Web Service (Status: %d)", resp.StatusCode)
		return service
	}

	if resp.StatusCode > 0 {
		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "unknown"
		}
		service.Name = fmt.Sprintf("HTTP Service (Port %d)", port)
		service.Type = "http"
		service.Description = fmt.Sprintf("HTTP Service (Status: %d, Content-Type: %s)", resp.StatusCode, contentType)
		return service
	}

	return nil
}

// guessServiceByPort guesses service type based on common port numbers (fallback only)
func (d *Detector) guessServiceByPort(port int) *Service {
	portMap := map[int]*Service{
		3000: {Name: "Node.js Dev Server", Type: "webapp", Description: "Common port for Node.js development servers"},
		3001: {Name: "Alternative Web Service", Type: "web", Description: "Common alternative port for web services"},
		8080: {Name: "HTTP Proxy/Web Server", Type: "web", Description: "Common HTTP alternative port"},
		8081: {Name: "HTTP Alternative", Type: "web", Description: "Common HTTP alternative port"},
		9090: {Name: "Prometheus", Type: "prometheus", Description: "Default Prometheus port"},
		9091: {Name: "Prometheus Alternative", Type: "prometheus", Description: "Alternative Prometheus port"},
		8000: {Name: "Python HTTP Server", Type: "web", Description: "Common Python development server port"},
		8001: {Name: "Python HTTP Alternative", Type: "web", Description: "Alternative Python server port"},
		5000: {Name: "Flask/Development Server", Type: "web", Description: "Common Flask development port"},
		5001: {Name: "Flask Alternative", Type: "web", Description: "Alternative Flask port"},
		4000: {Name: "Development Server", Type: "web", Description: "Common development server port"},
		7000: {Name: "Development Server", Type: "web", Description: "Common development server port"},
		9000: {Name: "SonarQube/Development", Type: "web", Description: "Common for SonarQube or development servers"},
		8888: {Name: "Jupyter Notebook", Type: "jupyter", Description: "Common Jupyter Notebook port"},
		5601: {Name: "Kibana", Type: "kibana", Description: "Default Kibana port"},
		9200: {Name: "Elasticsearch", Type: "elasticsearch", Description: "Default Elasticsearch port"},
		15672: {Name: "RabbitMQ Management", Type: "rabbitmq", Description: "RabbitMQ Management UI"},
		6379: {Name: "Redis", Type: "redis", Description: "Default Redis port"},
		27017: {Name: "MongoDB", Type: "mongodb", Description: "Default MongoDB port"},
		5432: {Name: "PostgreSQL", Type: "postgres", Description: "Default PostgreSQL port"},
		3306: {Name: "MySQL", Type: "mysql", Description: "Default MySQL port"},
	}

	if svc, exists := portMap[port]; exists {
		return &Service{
			Port:        port,
			Name:        svc.Name,
			Type:        svc.Type,
			URL:         fmt.Sprintf("http://localhost:%d", port),
			Description: svc.Description,
		}
	}

	return nil
}

