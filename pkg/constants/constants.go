package constants

const (
	// DefaultPortRange is the default port range to scan
	DefaultPortRange = "3000-9000"
	
	// DefaultDashboardPort is the default port for the web dashboard
	DefaultDashboardPort = 8080
	
	// DefaultTunnelStartPort is the default starting port for local tunnels
	DefaultTunnelStartPort = 9000
	
	// DefaultDetectionMode is the default service detection mode
	DefaultDetectionMode = "both"
	
	// DetectionModeDocker uses only Docker-based detection
	DetectionModeDocker = "docker"
	
	// DetectionModeDirect uses only direct port scanning
	DetectionModeDirect = "direct"
	
	// DetectionModeBoth uses both Docker and direct detection
	DetectionModeBoth = "both"
	
	// TunnelStabilizeDelay is the time to wait for tunnels to stabilize
	TunnelStabilizeDelay = 2
	
	// ServiceProbeTimeout is the timeout for service probing
	ServiceProbeTimeout = 3
	
	// MinValidPort is the minimum valid port number
	MinValidPort = 1024
	
	// MaxValidPort is the maximum valid port number
	MaxValidPort = 65535
	
	// NginxPreferredPorts are the preferred ports for Nginx Proxy Manager
	NginxPreferredPorts = "443,80,81"
)

