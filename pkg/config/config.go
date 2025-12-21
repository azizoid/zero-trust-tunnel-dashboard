package config

// Config holds all application configuration
type Config struct {
	SSH          SSHConfig
	Scan         ScanConfig
	Dashboard    DashboardConfig
	Detection    DetectionConfig
	Tunnel       TunnelConfig
}

// SSHConfig holds SSH connection configuration
type SSHConfig struct {
	Host         string
	Server       string
	User         string
	KeyPath      string
	UseHostAlias bool
	HostAlias    string
}

// ScanConfig holds port scanning configuration
type ScanConfig struct {
	PortRange string
}

// DashboardConfig holds dashboard server configuration
type DashboardConfig struct {
	Port int
}

// DetectionConfig holds service detection configuration
type DetectionConfig struct {
	Mode string
}

// TunnelConfig holds tunnel configuration
type TunnelConfig struct {
	StartPort int
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Scan: ScanConfig{
			PortRange: "3000-9000",
		},
		Dashboard: DashboardConfig{
			Port: 8080,
		},
		Detection: DetectionConfig{
			Mode: "both",
		},
		Tunnel: TunnelConfig{
			StartPort: 9000,
		},
	}
}

