package tunnel

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/ssh"
)

type Manager struct {
	server       string
	user         string
	keyPath      string
	useHostAlias bool
	hostAlias    string
	insecure     bool
	tunnels      map[int]*Tunnel
	tunnelsMu    sync.RWMutex
	localPorts   map[int]int
	portsMu      sync.RWMutex
	nextPort     int
	startPort    int
}

type Tunnel struct {
	RemotePort int
	LocalPort  int
	Cmd        *exec.Cmd
	ctx        context.Context
	cancel     context.CancelFunc
	errChan    chan error
}

func NewManager(server, user, keyPath string, startPort int) *Manager {
	return &Manager{
		server:       server,
		user:         user,
		keyPath:      keyPath,
		useHostAlias: false,
		tunnels:      make(map[int]*Tunnel),
		localPorts:   make(map[int]int),
		nextPort:     startPort,
		startPort:    startPort,
	}
}

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

func (m *Manager) SetInsecure(insecure bool) {
	m.insecure = insecure
}

func (m *Manager) CreateTunnel(remotePort int) (int, error) {
	m.tunnelsMu.Lock()
	defer m.tunnelsMu.Unlock()

	if tunnel, exists := m.tunnels[remotePort]; exists {
		// Check if existing tunnel is healthy
		select {
		case <-tunnel.errChan:
			// Tunnel dead, clean up and recreate
			m.cleanupTunnel(remotePort)
		default:
			if tunnel.Cmd.ProcessState == nil || !tunnel.Cmd.ProcessState.Exited() {
				return tunnel.LocalPort, nil
			}
			m.cleanupTunnel(remotePort)
		}
	}

	localPort := remotePort
	if localPort < 1024 || localPort > 65535 {
		localPort = m.nextPort
		m.nextPort++
	} else {
		if _, exists := m.localPorts[localPort]; exists {
			localPort = m.nextPort
			m.nextPort++
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Use ssh.Client to build the command with proper config
	sshConfig := ssh.Config{
		Server:       m.server,
		User:         m.user,
		KeyPath:      m.keyPath,
		UseHostAlias: m.useHostAlias,
		HostAlias:    m.hostAlias,
		Insecure:     m.insecure,
	}
	sshClient := ssh.NewClient(sshConfig)
	cmd := sshClient.BuildTunnelCommand(ctx, localPort, remotePort)

	tunnel := &Tunnel{
		RemotePort: remotePort,
		LocalPort:  localPort,
		Cmd:        cmd,
		ctx:        ctx,
		cancel:     cancel,
		errChan:    make(chan error, 1),
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return 0, fmt.Errorf("failed to start tunnel: %w", err)
	}

	// Start monitoring routine
	go m.monitorTunnel(tunnel)

	// Wait briefly to catch immediate failures (e.g. auth error, port in use)
	select {
	case err := <-tunnel.errChan:
		cancel()
		return 0, fmt.Errorf("tunnel failed immediately: %w", err)
	case <-time.After(500 * time.Millisecond):
		// Tunnel seems stable enough for now
	}

	m.tunnels[remotePort] = tunnel
	m.portsMu.Lock()
	m.localPorts[remotePort] = localPort
	m.portsMu.Unlock()

	return localPort, nil
}

func (m *Manager) monitorTunnel(t *Tunnel) {
	err := t.Cmd.Wait()
	if err != nil {
		// If context was canceled, it's an intentional stop
		if t.ctx.Err() != nil {
			return
		}
		t.errChan <- err
	} else if t.ctx.Err() == nil {
		// Process exited with 0, still means tunnel closed
		t.errChan <- fmt.Errorf("tunnel process exited unexpectedly with code 0")
	}
	close(t.errChan)
}

func (m *Manager) cleanupTunnel(remotePort int) {
	// Assumes caller holds lock
	if t, ok := m.tunnels[remotePort]; ok {
		delete(m.tunnels, remotePort)
		m.portsMu.Lock()
		delete(m.localPorts, remotePort)
		m.portsMu.Unlock()
		t.cancel() // Just in case
	}
}

// GetLocalPort returns the local port mapping for a given remote port.
func (m *Manager) GetLocalPort(remotePort int) (int, bool) {
	m.portsMu.RLock()
	defer m.portsMu.RUnlock()
	localPort, exists := m.localPorts[remotePort]
	return localPort, exists
}

// CloseTunnel closes a specific tunnel by remote port.
func (m *Manager) CloseTunnel(remotePort int) error {
	m.tunnelsMu.Lock()
	defer m.tunnelsMu.Unlock()

	tunnel, exists := m.tunnels[remotePort]
	if !exists {
		return nil
	}

	tunnel.cancel()
	// No need to kill explicitly, cancel context does it for exec.CommandContext

	delete(m.tunnels, remotePort)
	m.portsMu.Lock()
	delete(m.localPorts, remotePort)
	m.portsMu.Unlock()

	return nil
}

func (m *Manager) CloseAll() {
	m.tunnelsMu.Lock()
	defer m.tunnelsMu.Unlock()

	for _, tunnel := range m.tunnels {
		tunnel.cancel()
	}

	m.tunnels = make(map[int]*Tunnel)
	m.portsMu.Lock()
	m.localPorts = make(map[int]int)
	m.portsMu.Unlock()
}

// HealthCheck verifies if a tunnel is still active.
func (m *Manager) HealthCheck(remotePort int) bool {
	m.tunnelsMu.RLock()
	defer m.tunnelsMu.RUnlock()

	tunnel, exists := m.tunnels[remotePort]
	if !exists {
		return false
	}

	select {
	case <-tunnel.errChan:
		return false
	default:
		return true
	}
}
