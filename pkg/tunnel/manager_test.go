package tunnel

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager("example.com", "user", "/key", 9000)
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.server != "example.com" {
		t.Errorf("expected server example.com, got %s", m.server)
	}
}

func TestNewManagerWithHost(t *testing.T) {
	m := NewManagerWithHost("host-alias", 9000)
	if m == nil {
		t.Fatal("NewManagerWithHost returned nil")
	}
	if !m.useHostAlias {
		t.Error("expected useHostAlias to be true")
	}
}

func TestCleanupTunnel(t *testing.T) {
	m := NewManager("example.com", "user", "/key", 9000)

	// Manually inject a dummy tunnel structure
	// Note: We can't easily mock the exec.Cmd execution in this unit test without
	// significant refactoring to use an interface for command execution.
	// So we are testing the state management logic here.

	m.tunnelsMu.Lock()
	tunnel := &Tunnel{}
	m.tunnels[8080] = tunnel
	m.localPorts[8080] = 9000
	// Mock the cancel func to avoid nil pointer
	tunnel.cancel = func() {}
	m.tunnelsMu.Unlock()

	m.tunnelsMu.Lock()
	m.cleanupTunnel(8080)
	m.tunnelsMu.Unlock()

	if _, exists := m.tunnels[8080]; exists {
		t.Error("Tunnel should have been removed from tunnels map")
	}

	if _, exists := m.GetLocalPort(8080); exists {
		t.Error("Port mapping should have been removed")
	}
}
