package config

type Config struct {
	SSH       SSHConfig
	Scan      ScanConfig
	Dashboard DashboardConfig
	Detection DetectionConfig
	Tunnel    TunnelConfig
}

type SSHConfig struct {
	Host         string
	Server       string
	User         string
	KeyPath      string
	UseHostAlias bool
	HostAlias    string
}

type ScanConfig struct {
	PortRange string
}

type DashboardConfig struct {
	Port int
}

type DetectionConfig struct {
	Mode string
}

type TunnelConfig struct {
	StartPort int
}

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
