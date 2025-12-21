package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/azizoid/zero-trust-dashboard/pkg/dashboard"
	"github.com/azizoid/zero-trust-dashboard/pkg/detector"
	"github.com/azizoid/zero-trust-dashboard/pkg/scanner"
	"github.com/azizoid/zero-trust-dashboard/pkg/server"
	"github.com/azizoid/zero-trust-dashboard/pkg/sshconfig"
	"github.com/azizoid/zero-trust-dashboard/pkg/tunnel"
)

func main() {
	var (
		host            = flag.String("host", "", "SSH host alias from ~/.ssh/config (alternative to --server/--user)")
		serverAddr      = flag.String("server", "", "SSH server address (required if --host not set)")
		user            = flag.String("user", "", "SSH username (required if --host not set)")
		keyPath         = flag.String("key", "", "Path to SSH private key (optional, overrides SSH config)")
		scanPorts       = flag.String("scan-ports", "3000-9000", "Port range to scan (e.g., 3000-9000)")
		dashboardPort   = flag.Int("dashboard-port", 8080, "Port for the web dashboard")
		tunnelStartPort = flag.Int("tunnel-start-port", 9000, "Starting port for local tunnel ports")
		detectionMode   = flag.String("detection-mode", "both", "Service detection method: docker, direct, or both (default: both)")
	)
	flag.Parse()

	var finalServer, finalUser, finalKey string

	if *host != "" {
		sshConfig, err := sshconfig.ParseSSHConfig(*host)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading SSH config: %v\n", err)
			fmt.Fprintf(os.Stderr, "Make sure you have a Host entry for '%s' in ~/.ssh/config\n", *host)
			os.Exit(1)
		}

		finalServer = sshConfig.HostName
		finalUser = sshConfig.User
		finalKey = sshConfig.IdentityFile

		if *keyPath != "" {
			finalKey = *keyPath
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
		if *serverAddr == "" || *user == "" {
			fmt.Fprintf(os.Stderr, "Error: either --host or both --server and --user are required\n")
			fmt.Fprintf(os.Stderr, "  Use --host to read from ~/.ssh/config\n")
			fmt.Fprintf(os.Stderr, "  Or use --server and --user for direct connection\n")
			flag.Usage()
			os.Exit(1)
		}

		finalServer = *serverAddr
		finalUser = *user
		finalKey = *keyPath
	}

	fmt.Println("Zero-Trust Tunnel Dashboard")
	fmt.Println("===========================================================")
	if *host != "" {
		fmt.Printf("SSH Host: %s\n", *host)
	}
	fmt.Printf("Server: %s\n", finalServer)
	fmt.Printf("User: %s\n", finalUser)
	if finalKey != "" {
		fmt.Printf("Key: %s\n", finalKey)
	}
	fmt.Printf("Scanning ports: %s\n", *scanPorts)
	fmt.Println()

	var tunnelMgr *tunnel.Manager
	var portScanner *scanner.Scanner

	if *host != "" {
		tunnelMgr = tunnel.NewManagerWithHost(*host, *tunnelStartPort)
		portScanner = scanner.NewScannerWithHost(*host)
	} else {
		tunnelMgr = tunnel.NewManager(finalServer, finalUser, finalKey, *tunnelStartPort)
		portScanner = scanner.NewScanner(finalServer, finalUser, finalKey)
	}

	serviceDetector := detector.NewDetector(3 * time.Second)

	var ports []int
	var dockerServices map[int]*detector.DockerService
	var allContainers []*detector.DockerService

	if *detectionMode == "docker" || *detectionMode == "both" {
		var err error
		if *host != "" {
			dockerServices, err = detector.DetectDockerServices("", "", "", true, *host)
		} else {
			dockerServices, err = detector.DetectDockerServices(finalServer, finalUser, finalKey, false, "")
		}

		if err != nil {
			if *detectionMode == "docker" {
				fmt.Fprintf(os.Stderr, "Error: Docker detection failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Warning: Docker detection failed: %v\n", err)
			dockerServices = make(map[int]*detector.DockerService)
		}

		var err2 error
		if *host != "" {
			allContainers, err2 = detector.GetAllDockerContainers("", "", "", true, *host)
		} else {
			allContainers, err2 = detector.GetAllDockerContainers(finalServer, finalUser, finalKey, false, "")
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

	if *detectionMode == "direct" || (*detectionMode == "both" && len(ports) == 0) {
		fmt.Println("Scanning for open ports...")
		scannedPorts, err := portScanner.ScanPorts(*scanPorts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning ports: %v\n", err)
			os.Exit(1)
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
		os.Exit(0)
	}

	fmt.Printf("Found %d port(s) to tunnel: %v\n\n", len(ports), ports)

	fmt.Println("Creating SSH tunnels...")
	localPorts := make(map[int]int)
	for _, port := range ports {
		localPort, err := tunnelMgr.CreateTunnel(port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create tunnel for port %d: %v\n", port, err)
			continue
		}
		localPorts[port] = localPort
		fmt.Printf("   Tunnel created: localhost:%d -> %s:%d\n", localPort, finalServer, port)
	}

	if len(localPorts) == 0 {
		fmt.Fprintf(os.Stderr, "Failed to create any tunnels\n")
		os.Exit(1)
	}

	fmt.Println()

	fmt.Println("Waiting for tunnels to stabilize...")
	time.Sleep(2 * time.Second)

	fmt.Println("Detecting services...")

	if dockerServices == nil {
		dockerServices = make(map[int]*detector.DockerService)
	}

	useDirect := *detectionMode == "direct" || *detectionMode == "both"

	var services []detector.Service
	if useDirect {
		services = serviceDetector.DetectServices(ports, dockerServices)
	} else {
		services = serviceDetector.DetectServicesFromDocker(ports, dockerServices)
	}

	if *detectionMode == "docker" || *detectionMode == "both" {
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
						if *host != "" {
							domains, _ = detector.QueryNPMDatabase(nginxContainerName, container.ContainerName, container.Port, "", "", "", true, *host)
						} else {
							domains, _ = detector.QueryNPMDatabase(nginxContainerName, container.ContainerName, container.Port, finalServer, finalUser, finalKey, false, "")
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
					services = append(services, *service)
				}
			} else if !container.ExposedToHost {
				if !servicePortMap[container.Port] {
					service := detector.IdentifyServiceFromDocker(container)
					if service != nil {
						service.Network = container.Network
						if hasNginxProxy && nginxLocalPort > 0 && nginxContainerName != "" {
							var domains []string
							if *host != "" {
								domains, _ = detector.QueryNPMDatabase(nginxContainerName, container.ContainerName, container.Port, "", "", "", true, *host)
							} else {
								domains, _ = detector.QueryNPMDatabase(nginxContainerName, container.ContainerName, container.Port, finalServer, finalUser, finalKey, false, "")
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
						services = append(services, *service)
						servicePortMap[container.Port] = true
					}
				}
			} else {
			}
		}
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

	dashGen := dashboard.NewGenerator(services)
	html, err := dashGen.GenerateHTML(localPorts, *tunnelStartPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating dashboard: %v\n", err)
		os.Exit(1)
	}

	httpServer := server.NewServer(*dashboardPort, services)
	httpServer.SetHTML(html)
	httpServer.SetScanner(portScanner)

	go func() {
		if err := httpServer.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
			os.Exit(1)
		}
	}()

	fmt.Println(dashGen.GenerateCLI(localPorts))
	fmt.Printf("Web dashboard available at: http://localhost:%d\n", *dashboardPort)
	fmt.Println("\nPress Ctrl+C to stop...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n\nShutting down...")
	tunnelMgr.CloseAll()
	fmt.Println("All tunnels closed. Goodbye!")
}
