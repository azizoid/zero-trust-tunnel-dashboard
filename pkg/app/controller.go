package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/dashboard"
	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/detector"
	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/scanner"
	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/server"
	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/sshconfig"
	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/tunnel"
)

type Config struct {
	Host            string
	ServerAddr      string
	User            string
	KeyPath         string
	ScanPorts       string
	DashboardPort   int
	TunnelStartPort int
	DetectionMode   string
	Insecure        bool
}

type Controller struct {
	config      Config
	tunnelMgr   *tunnel.Manager
	portScanner *scanner.Scanner
	dashGen     *dashboard.Generator
	httpServer  *server.Server
	cancel      context.CancelFunc
}

func NewController(cfg Config) (*Controller, error) {
	return &Controller{
		config: cfg,
	}, nil
}

func (c *Controller) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	defer cancel()

	var finalServer, finalUser, finalKey string

	if c.config.Host != "" {
		sshConfig, err := sshconfig.ParseSSHConfig(c.config.Host)
		if err != nil {
			return fmt.Errorf("error reading SSH config: %v (make sure you have a Host entry for '%s')", err, c.config.Host)
		}

		finalServer = sshConfig.HostName
		finalUser = sshConfig.User
		finalKey = sshConfig.IdentityFile

		if c.config.KeyPath != "" {
			finalKey = c.config.KeyPath
		}

		if finalUser == "" {
			finalUser = os.Getenv("USER")
			if finalUser == "" {
				finalUser = os.Getenv("USERNAME")
				if finalUser == "" {
					finalUser = "root" // fallback
				}
			}
		}
	} else {
		finalServer = c.config.ServerAddr
		finalUser = c.config.User
		finalKey = c.config.KeyPath
	}

	fmt.Println("Zero-Trust Tunnel Dashboard")
	fmt.Println("===========================================================")
	if c.config.Host != "" {
		fmt.Printf("SSH Host: %s\n", c.config.Host)
	}
	fmt.Printf("Server: %s\n", finalServer)
	fmt.Printf("User: %s\n", finalUser)
	if finalKey != "" {
		fmt.Printf("Key: %s\n", finalKey)
	}
	if c.config.Insecure {
		fmt.Println("WARNING: Strict host key checking disabled!")
	}
	fmt.Printf("Scanning ports: %s\n", c.config.ScanPorts)
	fmt.Println()

	if c.config.Host != "" {
		c.tunnelMgr = tunnel.NewManagerWithHost(c.config.Host, c.config.TunnelStartPort)
		c.portScanner = scanner.NewScannerWithHost(c.config.Host)
	} else {
		c.tunnelMgr = tunnel.NewManager(finalServer, finalUser, finalKey, c.config.TunnelStartPort)
		c.portScanner = scanner.NewScanner(finalServer, finalUser, finalKey)
	}

	c.tunnelMgr.SetInsecure(c.config.Insecure)
	c.portScanner.SetInsecure(c.config.Insecure)

	serviceDetector := detector.NewDetector(3 * time.Second)

	var ports []int
	var dockerServices map[int]*detector.DockerService
	var allContainers []*detector.DockerService

	if c.config.DetectionMode == "docker" || c.config.DetectionMode == "both" {
		var err error
		if c.config.Host != "" {
			dockerServices, err = detector.DetectDockerServices("", "", "", true, c.config.Host, c.config.Insecure)
		} else {
			dockerServices, err = detector.DetectDockerServices(finalServer, finalUser, finalKey, false, "", c.config.Insecure)
		}

		if err != nil {
			if c.config.DetectionMode == "docker" {
				return fmt.Errorf("docker detection failed: %v", err)
			}
			fmt.Printf("Warning: Docker detection failed: %v\n", err)
			dockerServices = make(map[int]*detector.DockerService)
		}

		var err2 error
		if c.config.Host != "" {
			allContainers, err2 = detector.GetAllDockerContainers("", "", "", true, c.config.Host, c.config.Insecure)
		} else {
			allContainers, err2 = detector.GetAllDockerContainers(finalServer, finalUser, finalKey, false, "", c.config.Insecure)
		}

		if err2 == nil {
			totalContainers := len(allContainers)
			accessibleContainers := len(dockerServices)

			fmt.Printf("Found %d Docker container(s) (%d with exposed ports)\n", totalContainers, accessibleContainers)

			if totalContainers > accessibleContainers {
				fmt.Printf("Note: %d container(s) have no exposed ports or are on internal networks\n", totalContainers-accessibleContainers)
			}

			for port := range dockerServices {
				ports = append(ports, port)
			}
		} else if len(dockerServices) > 0 {
			fmt.Printf("Found %d Docker container(s) with exposed ports\n", len(dockerServices))
			for port := range dockerServices {
				ports = append(ports, port)
			}
		}
	}

	if c.config.DetectionMode == "direct" || (c.config.DetectionMode == "both" && len(ports) == 0) {
		fmt.Println("Scanning for open ports...")
		scannedPorts, err := c.portScanner.ScanPorts(c.config.ScanPorts)
		if err != nil {
			return fmt.Errorf("error scanning ports: %v", err)
		}

		portMap := make(map[int]bool)
		for _, p := range ports {
			portMap[p] = true
		}
		for _, p := range scannedPorts {
			if !portMap[p] {
				ports = append(ports, p)
			}
		}
	}

	if len(ports) == 0 {
		fmt.Println("No ports found to tunnel")
		return nil
	}

	fmt.Printf("Found %d port(s) to tunnel: %v\n\n", len(ports), ports)

	fmt.Println("Creating SSH tunnels...")
	localPorts := make(map[int]int)
	for _, port := range ports {
		localPort, err := c.tunnelMgr.CreateTunnel(port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create tunnel for port %d: %v\n", port, err)
			continue
		}
		localPorts[port] = localPort
		fmt.Printf("   Tunnel created: localhost:%d -> %s:%d\n", localPort, finalServer, port)
	}

	if len(localPorts) == 0 {
		return fmt.Errorf("failed to create any tunnels")
	}

	fmt.Println()

	fmt.Println("Waiting for tunnels to stabilize...")
	time.Sleep(2 * time.Second)

	fmt.Println("Detecting services...")

	if dockerServices == nil {
		dockerServices = make(map[int]*detector.DockerService)
	}

	useDirect := c.config.DetectionMode == "direct" || c.config.DetectionMode == "both"

	var services []detector.Service
	if useDirect {
		services = serviceDetector.DetectServices(ports, dockerServices)
	} else {
		services = serviceDetector.DetectServicesFromDocker(ports, dockerServices)
	}

	// Nginx Proxy Manager handling
	if c.config.DetectionMode == "docker" || c.config.DetectionMode == "both" {
		c.handleNginxProxy(services, dockerServices, allContainers, localPorts, finalServer, finalUser, finalKey, &services)
	}

	fmt.Printf("Detected %d service(s)\n\n", len(services))
	for i := range services {
		if services[i].Port > 0 {
			if strings.HasPrefix(services[i].URL, "https://") {
				services[i].URL = fmt.Sprintf("https://localhost:%d", services[i].Port)
			} else if services[i].URL == "" {
				services[i].URL = fmt.Sprintf("http://localhost:%d", services[i].Port)
			}
		}
	}

	c.dashGen = dashboard.NewGenerator(services)
	html, err := c.dashGen.GenerateHTML(localPorts, c.config.TunnelStartPort)
	if err != nil {
		return fmt.Errorf("error generating dashboard: %v", err)
	}

	c.httpServer = server.NewServer(c.config.DashboardPort, services)
	c.httpServer.SetHTML(html)
	c.httpServer.SetScanner(c.portScanner)

	go func() {
		if err := c.httpServer.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
			c.cancel() // Signal shutdown on server error
		}
	}()

	fmt.Println(c.dashGen.GenerateCLI(localPorts, c.config.TunnelStartPort))
	fmt.Printf("Web dashboard available at: http://localhost:%d\n", c.config.DashboardPort)
	fmt.Println("\nPress Ctrl+C to stop...")

	<-ctx.Done()
	fmt.Println("\n\nShutting down...")
	c.tunnelMgr.CloseAll()
	fmt.Println("All tunnels closed. Goodbye!")
	return nil
}

func (c *Controller) handleNginxProxy(services []detector.Service, dockerServices map[int]*detector.DockerService, allContainers []*detector.DockerService, localPorts map[int]int, server, user, key string, servicesPtr *[]detector.Service) {
	hasNginxProxy := false
	nginxRemotePort := 0

	var nginxPorts []int
	for port, dockerSvc := range dockerServices {
		if dockerSvc != nil && (strings.Contains(strings.ToLower(dockerSvc.Image), "nginx") || strings.Contains(strings.ToLower(dockerSvc.ContainerName), "nginx")) {
			hasNginxProxy = true
			nginxPorts = append(nginxPorts, port)
		}
	}

	if hasNginxProxy && len(nginxPorts) > 0 {
		for _, preferredPort := range []int{443, 80, 81} {
			for _, port := range nginxPorts {
				if port == preferredPort {
					nginxRemotePort = port
					break
				}
			}
			if nginxRemotePort > 0 {
				break
			}
		}

		if nginxRemotePort == 0 {
			nginxRemotePort = nginxPorts[0]
		}
	}

	if !hasNginxProxy {
		for _, svc := range services {
			if svc.Type == "nginx" || strings.Contains(strings.ToLower(svc.Name), "nginx") {
				hasNginxProxy = true
				nginxRemotePort = svc.Port
				break
			}
		}
	}

	nginxLocalPort := 0
	nginxContainerName := ""
	if nginxRemotePort > 0 {
		if lp, exists := localPorts[nginxRemotePort]; exists {
			nginxLocalPort = lp
		}
		for port, dockerSvc := range dockerServices {
			if port == nginxRemotePort && dockerSvc != nil {
				nginxContainerName = dockerSvc.ContainerName
				break
			}
		}
		if nginxContainerName == "" {
			for _, svc := range services {
				if svc.Type == "nginx" || strings.Contains(strings.ToLower(svc.Name), "nginx") {
					nginxContainerName = svc.Name
					break
				}
			}
		}
	}

	servicePortMap := make(map[int]bool)
	for _, svc := range services {
		if svc.Port > 0 {
			servicePortMap[svc.Port] = true
		}
	}

	for _, container := range allContainers {
		if !container.HasPorts || container.Port == 0 {
			service := detector.IdentifyServiceFromDocker(container)
			if service != nil {
				service.Network = container.Network
				if hasNginxProxy && nginxLocalPort > 0 && nginxContainerName != "" {
					var domains []string
					if c.config.Host != "" {
						domains, _ = detector.QueryNPMDatabase(nginxContainerName, container.ContainerName, container.Port, "", "", "", true, c.config.Host, c.config.Insecure) //nolint:errcheck
					} else {
						domains, _ = detector.QueryNPMDatabase(nginxContainerName, container.ContainerName, container.Port, server, user, key, false, "", c.config.Insecure) //nolint:errcheck
					}

					if len(domains) > 0 {
						domain := domains[0]
						service.Port = 0
						service.URL = fmt.Sprintf("http://localhost:%d", nginxLocalPort)
						service.Domain = domain
						service.Description = fmt.Sprintf("%s (Domain: %s)", service.Description, domain)
					} else {
						service.Port = 0
						service.URL = fmt.Sprintf("http://localhost:%d", nginxLocalPort)
						service.Description = fmt.Sprintf("%s (Accessible via Nginx Proxy Manager)", service.Description)
					}
				} else {
					service.Port = 0
					service.URL = ""
					service.Description = fmt.Sprintf("%s (No exposed ports - internal network only)", service.Description)
				}
				*servicesPtr = append(*servicesPtr, *service)
			}
		} else if !container.ExposedToHost {
			if !servicePortMap[container.Port] {
				service := detector.IdentifyServiceFromDocker(container)
				if service != nil {
					service.Network = container.Network
					if hasNginxProxy && nginxLocalPort > 0 && nginxContainerName != "" {
						var domains []string
						if c.config.Host != "" {
							domains, _ = detector.QueryNPMDatabase(nginxContainerName, container.ContainerName, container.Port, "", "", "", true, c.config.Host, c.config.Insecure) //nolint:errcheck
						} else {
							domains, _ = detector.QueryNPMDatabase(nginxContainerName, container.ContainerName, container.Port, server, user, key, false, "", c.config.Insecure) //nolint:errcheck
						}

						if len(domains) > 0 {
							domain := domains[0]
							service.Port = 0
							service.URL = fmt.Sprintf("http://localhost:%d", nginxLocalPort)
							service.Domain = domain
							service.Description = fmt.Sprintf("%s (Domain: %s)", service.Description, domain)
						} else {
							service.Port = 0
							service.URL = fmt.Sprintf("http://localhost:%d", nginxLocalPort)
							service.Description = fmt.Sprintf("%s (Accessible via Nginx Proxy Manager)", service.Description)
						}
					} else {
						service.Port = 0
						service.URL = ""
						service.Description = fmt.Sprintf("%s (Container port %d - not exposed to host)", service.Description, container.Port)
					}
					*servicesPtr = append(*servicesPtr, *service)
					servicePortMap[container.Port] = true
				}
			}
		}
	}
}
