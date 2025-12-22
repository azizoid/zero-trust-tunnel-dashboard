package config

import "testing"

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.Scan.PortRange != "3000-9000" {
		t.Errorf("Expected PortRange to be '3000-9000', got '%s'", config.Scan.PortRange)
	}

	if config.Dashboard.Port != 8080 {
		t.Errorf("Expected Dashboard Port to be 8080, got %d", config.Dashboard.Port)
	}

	if config.Detection.Mode != "both" {
		t.Errorf("Expected Detection Mode to be 'both', got '%s'", config.Detection.Mode)
	}

	if config.Tunnel.StartPort != 9000 {
		t.Errorf("Expected Tunnel StartPort to be 9000, got %d", config.Tunnel.StartPort)
	}
}
