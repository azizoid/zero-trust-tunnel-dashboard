package tunnel

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// Manager handles SSH tunnel lifecycle
type Manager struct {
	server       string
	user         string
	keyPath      string
	useHostAlias bool
	hostAlias    string
	tunnels    map[int]*Tunnel
	tunnelsMu  sync.RWMutex
	localPorts map[int]int
	portsMu    sync.RWMutex
	nextPort   int
	startPort  int
}

// Tunnel represents a single SSH tunnel
type Tunnel struct {
	RemotePort int
	LocalPort  int
	Cmd        *exec.Cmd
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewManager creates a new tunnel manager
func NewManager(server, user, keyPath string, startPort int) *Manager {
	return &Manager{
		server:     server,
		user:       user,
		keyPath:    keyPath,
		useHostAlias: false,
		tunnels:    make(map[int]*Tunnel),
		localPorts: make(map[int]int),
		nextPort:   startPort,
		startPort:  startPort,
	}
}

// NewManagerWithHost creates a new tunnel manager using SSH config host alias
func NewManagerWithHost(hostAlias string, startPort int) *Manager {
	return &Manager{
		useHostAlias: true,
		hostAlias:    hostAlias,
		tunnels:      make(map[int]*Tunnel),
		localPorts:   make(map[int]int),
		nextPort:     startPort,
		startPort:    startPort,
	}
}

// CreateTunnel creates an SSH tunnel for a remote port
func (m *Manager) CreateTunnel(remotePort int) (int, error) {
	m.tunnelsMu.Lock()
	defer m.tunnelsMu.Unlock()

	// Check if tunnel already exists
	if tunnel, exists := m.tunnels[remotePort]; exists {
		return tunnel.LocalPort, nil
	}

	// Use the remote port as the local port if available, otherwise use nextPort
	localPort := remotePort
	if localPort < 1024 || localPort > 65535 {
		// If port is out of range, use nextPort
		localPort = m.nextPort
		m.nextPort++
	} else {
		// Check if local port is already in use
		if _, exists := m.localPorts[localPort]; exists {
			localPort = m.nextPort
			m.nextPort++
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	args := []string{
		"-L", fmt.Sprintf("%d:localhost:%d", localPort, remotePort),
		"-N",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
	}

	if m.useHostAlias {
		args = append(args, m.hostAlias)
	} else {
		if m.keyPath != "" {
			args = append(args, "-i", m.keyPath)
		}
		args = append(args, fmt.Sprintf("%s@%s", m.user, m.server))
	}

	cmd := exec.CommandContext(ctx, "ssh", args...)

	tunnel := &Tunnel{
		RemotePort: remotePort,
		LocalPort:  localPort,
		Cmd:        cmd,
		ctx:        ctx,
		cancel:     cancel,
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return 0, fmt.Errorf("failed to start tunnel: %w", err)
	}

	time.Sleep(500 * time.Millisecond)
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		cancel()
		return 0, fmt.Errorf("tunnel failed to start")
	}

	m.tunnels[remotePort] = tunnel
	m.portsMu.Lock()
	m.localPorts[remotePort] = localPort
	m.portsMu.Unlock()

	return localPort, nil
}

// GetLocalPort returns the local port for a remote port
func (m *Manager) GetLocalPort(remotePort int) (int, bool) {
	m.portsMu.RLock()
	defer m.portsMu.RUnlock()
	localPort, exists := m.localPorts[remotePort]
	return localPort, exists
}

// CloseTunnel closes a specific tunnel
func (m *Manager) CloseTunnel(remotePort int) error {
	m.tunnelsMu.Lock()
	defer m.tunnelsMu.Unlock()

	tunnel, exists := m.tunnels[remotePort]
	if !exists {
		return nil
	}

	tunnel.cancel()
	if tunnel.Cmd.Process != nil {
		tunnel.Cmd.Process.Kill()
	}

	delete(m.tunnels, remotePort)
	m.portsMu.Lock()
	delete(m.localPorts, remotePort)
	m.portsMu.Unlock()

	return nil
}

// CloseAll closes all tunnels
func (m *Manager) CloseAll() {
	m.tunnelsMu.Lock()
	defer m.tunnelsMu.Unlock()

	for _, tunnel := range m.tunnels {
		tunnel.cancel()
		if tunnel.Cmd.Process != nil {
			tunnel.Cmd.Process.Kill()
		}
	}

	m.tunnels = make(map[int]*Tunnel)
	m.portsMu.Lock()
	m.localPorts = make(map[int]int)
	m.portsMu.Unlock()
}

// HealthCheck verifies if a tunnel is still running
func (m *Manager) HealthCheck(remotePort int) bool {
	m.tunnelsMu.RLock()
	defer m.tunnelsMu.RUnlock()

	tunnel, exists := m.tunnels[remotePort]
	if !exists {
		return false
	}

	if tunnel.Cmd.Process == nil {
		return false
	}

	return tunnel.Cmd.ProcessState == nil || !tunnel.Cmd.ProcessState.Exited()
}

