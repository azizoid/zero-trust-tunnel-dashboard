package detector

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDetector(t *testing.T) {
	detector := NewDetector(5 * time.Second)
	if detector == nil {
		t.Fatal("NewDetector() returned nil")
	}

	if detector.timeout != 5*time.Second {
		t.Errorf("Expected timeout to be 5s, got %v", detector.timeout)
	}
}

func TestNewDetectorWithZeroTimeout(t *testing.T) {
	detector := NewDetector(0)
	if detector == nil {
		t.Fatal("NewDetector() returned nil")
	}

	if detector.timeout != 3*time.Second {
		t.Errorf("Expected default timeout to be 3s, got %v", detector.timeout)
	}
}

func TestGuessServiceByPort(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		wantType string
		wantName string
	}{
		{
			name:     "Grafana port",
			port:     3000,
			wantType: "webapp",
			wantName: "Node.js Dev Server",
		},
		{
			name:     "Prometheus port",
			port:     9090,
			wantType: "prometheus",
			wantName: "Prometheus",
		},
		{
			name:     "Jupyter port",
			port:     8888,
			wantType: "jupyter",
			wantName: "Jupyter Notebook",
		},
		{
			name:     "Unknown port",
			port:     12345,
			wantType: "",
			wantName: "",
		},
	}

	detector := NewDetector(3 * time.Second)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := detector.guessServiceByPort(tt.port)
			if tt.wantType == "" {
				if service != nil {
					t.Errorf("guessServiceByPort() should return nil for port %d, got %v", tt.port, service)
				}
				return
			}

			if service == nil {
				t.Fatalf("guessServiceByPort() returned nil for port %d", tt.port)
			}

			if service.Type != tt.wantType {
				t.Errorf("guessServiceByPort() Type = %v, want %v", service.Type, tt.wantType)
			}

			if service.Name != tt.wantName {
				t.Errorf("guessServiceByPort() Name = %v, want %v", service.Name, tt.wantName)
			}
		})
	}
}

func TestIdentifyServiceFromResponse_Grafana(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Grafana-Version", "9.5.0")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	detector := NewDetector(3 * time.Second)
	client := &http.Client{Timeout: 3 * time.Second}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	service := detector.identifyServiceFromResponse(resp, 3000, "http")

	if service == nil {
		t.Fatal("identifyServiceFromResponse() returned nil")
	}

	if service.Type != "grafana" {
		t.Errorf("Expected Type to be 'grafana', got '%s'", service.Type)
	}

	if service.Name != "Grafana" {
		t.Errorf("Expected Name to be 'Grafana', got '%s'", service.Name)
	}
}

func TestIdentifyServiceFromResponse_Prometheus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# HELP test_metric\n# TYPE test_metric counter"))
	}))
	defer server.Close()

	detector := NewDetector(3 * time.Second)
	client := &http.Client{Timeout: 3 * time.Second}

	req, _ := http.NewRequest("GET", server.URL+"/metrics", nil)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	service := detector.identifyServiceFromResponse(resp, 9090, "http")

	if service == nil {
		t.Fatal("identifyServiceFromResponse() returned nil")
	}

	if service.Type != "prometheus" {
		t.Errorf("Expected Type to be 'prometheus', got '%s'", service.Type)
	}
}

func TestIdentifyServiceFromResponse_JSONAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	detector := NewDetector(3 * time.Second)
	client := &http.Client{Timeout: 3 * time.Second}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	service := detector.identifyServiceFromResponse(resp, 8080, "http")

	if service == nil {
		t.Fatal("identifyServiceFromResponse() returned nil")
	}

	if service.Type != "api" {
		t.Errorf("Expected Type to be 'api', got '%s'", service.Type)
	}
}

func TestDetectServicesFromDocker(t *testing.T) {
	detector := NewDetector(3 * time.Second)

	dockerServices := map[int]*DockerService{
		3000: {
			ContainerName: "grafana",
			Image:         "grafana/grafana",
			Port:          3000,
			HasPorts:      true,
		},
		8080: {
			ContainerName: "webapp",
			Image:         "nginx",
			Port:          8080,
			HasPorts:      true,
		},
	}

	ports := []int{3000, 8080, 9999}
	services := detector.DetectServicesFromDocker(ports, dockerServices)

	if len(services) != 3 {
		t.Errorf("Expected 3 services, got %d", len(services))
	}

	foundGrafana := false
	foundNginx := false
	foundUnknown := false

	for _, svc := range services {
		if svc.Type == "grafana" {
			foundGrafana = true
		}
		if svc.Type == "nginx" && svc.Port == 8080 {
			foundNginx = true
		}
		if svc.Type == "unknown" && svc.Port == 9999 {
			foundUnknown = true
		}
	}

	if !foundGrafana {
		t.Error("Grafana service not detected")
	}

	if !foundNginx {
		t.Error("Nginx service not detected")
	}

	if !foundUnknown {
		t.Error("Unknown service not detected for port without Docker info")
	}
}

func TestIdentifyServiceFromDocker(t *testing.T) {
	tests := []struct {
		name        string
		dockerSvc   *DockerService
		wantType    string
		wantName    string
		wantNetwork string
	}{
		{
			name: "Grafana container",
			dockerSvc: &DockerService{
				ContainerName: "grafana",
				Image:         "grafana/grafana:latest",
				Port:          3000,
				Network:       "bridge",
				HasPorts:      true,
			},
			wantType:    "grafana",
			wantName:    "Grafana",
			wantNetwork: "bridge",
		},
		{
			name: "Prometheus container",
			dockerSvc: &DockerService{
				ContainerName: "prometheus",
				Image:         "prom/prometheus",
				Port:          9090,
				Network:       "monitoring",
				HasPorts:      true,
			},
			wantType:    "prometheus",
			wantName:    "Prometheus",
			wantNetwork: "monitoring",
		},
		{
			name: "Generic application",
			dockerSvc: &DockerService{
				ContainerName: "my-app",
				Image:         "myapp:1.0",
				Port:          8080,
				Network:       "default",
				HasPorts:      true,
			},
			wantType:    "application",
			wantName:    "my-app",
			wantNetwork: "default",
		},
		{
			name: "Generic Docker container",
			dockerSvc: &DockerService{
				ContainerName: "some-container",
				Image:         "ubuntu:20.04",
				Port:          5000,
				Network:       "bridge",
				HasPorts:      true,
			},
			wantType:    "docker",
			wantName:    "some-container",
			wantNetwork: "bridge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := IdentifyServiceFromDocker(tt.dockerSvc)
			if service == nil {
				t.Fatal("IdentifyServiceFromDocker() returned nil")
			}

			if service.Type != tt.wantType {
				t.Errorf("IdentifyServiceFromDocker() Type = %v, want %v", service.Type, tt.wantType)
			}

			if service.Name != tt.wantName {
				t.Errorf("IdentifyServiceFromDocker() Name = %v, want %v", service.Name, tt.wantName)
			}

			if service.Network != tt.wantNetwork {
				t.Errorf("IdentifyServiceFromDocker() Network = %v, want %v", service.Network, tt.wantNetwork)
			}

			if service.Port != tt.dockerSvc.Port {
				t.Errorf("IdentifyServiceFromDocker() Port = %v, want %v", service.Port, tt.dockerSvc.Port)
			}
		})
	}
}

