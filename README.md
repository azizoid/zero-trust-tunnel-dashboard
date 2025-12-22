# Zero-Trust Tunnel Dashboard

[![CI](https://github.com/azizoid/zero-trust-tunnel-dashboard/workflows/CI/badge.svg)](https://github.com/azizoid/zero-trust-tunnel-dashboard/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/azizoid/zero-trust-tunnel-dashboard)](https://goreportcard.com/report/github.com/azizoid/zero-trust-tunnel-dashboard)
[![codecov](https://codecov.io/gh/azizoid/zero-trust-tunnel-dashboard/branch/main/graph/badge.svg)](https://codecov.io/gh/azizoid/zero-trust-tunnel-dashboard)

A Go-based CLI tool that automates SSH tunnel creation, port scanning, service detection, and provides a beautiful dashboard for accessing closed services (like Grafana, Prometheus, etc.) on remote servers through zero-trust tunnels.

## Problem Statement

Accessing services running on remote servers (like Grafana, Prometheus, or custom web apps) typically requires:
- Opening firewall ports (security risk)
- Setting up VPNs (complexity overhead)
- Using cloud-specific solutions (vendor lock-in)
- Manual SSH tunnel management (tedious and error-prone)

This tool solves these problems by providing **automated zero-trust tunnel access** - services remain behind firewalls, but are easily accessible through encrypted SSH tunnels with automatic service discovery.

## Features

- **Automatic Port Scanning**: Scans remote server ports via SSH
- **SSH Tunnel Management**: Automatically creates SSH tunnels for detected ports
- **Service Detection**: Identifies services by probing HTTP/HTTPS endpoints
  - Grafana
  - Prometheus
  - Kubernetes Dashboard
  - Jenkins
  - Jupyter Notebooks
  - Generic Web Services
  - REST APIs
- **Web Dashboard**: Beautiful, modern web interface to access all services
- **CLI Interface**: Terminal-friendly output with service information
- **Zero-Trust Access**: Secure access to services through SSH tunnels

## Prerequisites

- Go 1.23 or later
- SSH access to the target server
- `ss` or `netstat` command available on the remote server

## SSH Config Setup

To use the `--host` option, add an entry to your `~/.ssh/config` file:

```
Host my-server
    HostName example.com
    User admin
    IdentityFile ~/.ssh/id_rsa
    Port 22
```

Then you can simply run:

```bash
./tunnel-dash --host my-server
```

## Quick Start

```bash
# Clone the repository
git clone https://github.com/azizoid/zero-trust-tunnel-dashboard.git
cd zero-trust-tunnel-dashboard

# Build the tool
go build -o tunnel-dash ./cmd/tunnel-dash

# Run it (using SSH config)
./tunnel-dash --host your-server

# Or with direct connection
./tunnel-dash --server example.com --user admin
```

## Installation

### Prerequisites

- Go 1.21 or later ([install Go](https://go.dev/dl/))
- **Recommended**: Go 1.24+ for security (fixes standard library vulnerabilities in Go 1.23.x)
- SSH access to your target server
- `ss` or `netstat` command available on the remote server

### Build from Source

**Using Make (Recommended):**
```bash
# Clone the repository
git clone https://github.com/azizoid/zero-trust-tunnel-dashboard.git
cd zero-trust-tunnel-dashboard

# Build the tool
make build

# The binary is now ready to use
./tunnel-dash --help
```

**Using Go directly:**
```bash
# Clone the repository
git clone https://github.com/azizoid/zero-trust-tunnel-dashboard.git
cd zero-trust-tunnel-dashboard

# Build the tool
go build -o tunnel-dash ./cmd/tunnel-dash

# The binary is now ready to use
./tunnel-dash --help
```

**Reproducible Build:**
```bash
# Build with reproducible flags (deterministic binary)
make build-reproducible
```

### Build with Version Information

**Using Make:**
```bash
# Build with automatic version detection
make build

# Or specify version explicitly
VERSION=v0.1.0 make build
```

**Using Go directly:**
```bash
VERSION="v0.1.0"
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

go build -ldflags "-X github.com/azizoid/zero-trust-tunnel-dashboard/pkg/version.Version=${VERSION} -X github.com/azizoid/zero-trust-tunnel-dashboard/pkg/version.Commit=${COMMIT} -X github.com/azizoid/zero-trust-tunnel-dashboard/pkg/version.BuildDate=${BUILD_DATE}" -o tunnel-dash ./cmd/tunnel-dash

# Verify version
./tunnel-dash --version
```

### Development Commands

The project includes a `Makefile` with common development tasks:

```bash
make test        # Run tests
make bench       # Run benchmarks
make lint        # Run linter
make vulncheck   # Check for vulnerabilities
make clean       # Clean build artifacts
make help        # Show all available commands
```

## Usage

### Using SSH Config (Recommended)

If you have your SSH connection configured in `~/.ssh/config`, you can use the host alias directly:

```bash
./tunnel-dash --host my-server
```

This will automatically read the server address, user, and key from your SSH config file.

### Direct Connection

```bash
./tunnel-dash --server example.com --user admin
```

### With SSH Key

```bash
./tunnel-dash --server example.com --user admin --key ~/.ssh/id_rsa
```

Or with SSH config (key is read from config):

```bash
./tunnel-dash --host my-server
```

### Custom Port Range

```bash
./tunnel-dash --host my-server --scan-ports 3000-9000
```

### Custom Dashboard Port

```bash
./tunnel-dash --host my-server --dashboard-port 8080
```

### Detection Mode Examples

```bash
# Use Docker detection only (fastest, requires Docker)
./tunnel-dash --host my-server --detection-mode docker

# Use HTTP probing only (works without Docker)
./tunnel-dash --host my-server --detection-mode direct

# Use both methods (default, most accurate)
./tunnel-dash --host my-server --detection-mode both
```

### All Options

```bash
./tunnel-dash \
  --host my-server \
  --scan-ports 3000-9000 \
  --dashboard-port 8080 \
  --tunnel-start-port 9000 \
  --detection-mode both
```

Or with direct connection:

```bash
./tunnel-dash \
  --server example.com \
  --user admin \
  --key ~/.ssh/id_rsa \
  --scan-ports 3000-9000 \
  --dashboard-port 8080 \
  --tunnel-start-port 9000 \
  --detection-mode both
```

## Command-Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--host` | SSH host alias from ~/.ssh/config (alternative to --server/--user) | - |
| `--server` | SSH server address (required if --host not set) | - |
| `--user` | SSH username (required if --host not set) | - |
| `--key` | Path to SSH private key (optional, overrides SSH config) | - |
| `--scan-ports` | Port range to scan (e.g., 3000-9000) | 3000-9000 |
| `--dashboard-port` | Port for the web dashboard | 8080 |
| `--tunnel-start-port` | Starting port for local tunnel ports | 9000 |
| `--detection-mode` | Service detection method: `docker`, `direct`, or `both` | both |
| `--version` | Show version information and exit | - |

**Note**: Either use `--host` (reads from SSH config) or use both `--server` and `--user` (direct connection).

### Detection Modes

- **`both`** (default): Uses Docker container information and HTTP probing for best accuracy
- **`docker`**: Only uses Docker container information (faster, requires Docker on remote server)
- **`direct`**: Only uses HTTP probing (works without Docker, may be slower)

## How It Works

1. **Port Scanning**: The tool connects to the remote server via SSH and executes `ss -tlnp` or `netstat -tlnp` to find listening ports
2. **Tunnel Creation**: For each detected port, an SSH tunnel is created using `ssh -L`
3. **Service Detection**: The tool probes each port via HTTP/HTTPS to identify the service type
4. **Dashboard Generation**: A web dashboard is generated with links to all detected services
5. **Access**: Services are accessible through the local tunnel ports

## Example Output

```
Zero-Trust Tunnel Dashboard
===========================================================
Server: example.com
User: admin
Scanning ports: 3000-9000

Scanning for open ports...
Found 3 open port(s): [3000 9090 8080]

Creating SSH tunnels...
   Tunnel created: localhost:9000 -> example.com:3000
   Tunnel created: localhost:9001 -> example.com:9090
   Tunnel created: localhost:9002 -> example.com:8080

Waiting for tunnels to stabilize...
Detecting services...
Detected 3 service(s)

Zero-Trust Tunnel Dashboard
===========================================================

Grafana
   Type: grafana
   Description: Grafana Dashboard (Version: 9.5.0)
   Remote Port: 3000
   Local Port: 9000
   URL: http://localhost:9000

Prometheus
   Type: prometheus
   Description: Prometheus Metrics Server
   Remote Port: 9090
   Local Port: 9001
   URL: http://localhost:9001

Web Service (Port 8080)
   Type: web
   Description: Generic Web Service
   Remote Port: 8080
   Local Port: 9002
   URL: http://localhost:9002

Web dashboard available at: http://localhost:8080

Press Ctrl+C to stop...
```

## Architecture

The tool consists of several components:

- **Tunnel Manager** (`pkg/tunnel`): Manages SSH tunnel lifecycle
- **Port Scanner** (`pkg/scanner`): Scans remote ports via SSH
- **Service Detector** (`pkg/detector`): Identifies services by probing ports
- **Dashboard Generator** (`pkg/dashboard`): Generates HTML and CLI output
- **HTTP Server** (`pkg/server`): Serves the web dashboard

## Security Model

This tool implements a **zero-trust tunnel** approach for secure access to remote services.

### Zero-Trust Principles

1. **No Trust in Network**: All communication goes through encrypted SSH tunnels
2. **No Direct Exposure**: Remote services are never exposed to the internet
3. **Local Access Only**: Services are accessed through localhost-only tunnels
4. **SSH Authentication Required**: Access requires valid SSH credentials

### Security Features

- **Encrypted Tunnels**: All traffic is encrypted via SSH
- **Localhost Binding**: Tunnels and dashboard bind to `127.0.0.1` only
- **No Network Exposure**: Remote services remain behind firewall
- **SSH Key Authentication**: Uses standard SSH key-based authentication

### Security Considerations

⚠️ **SSH Host Key Verification**: The tool disables `StrictHostKeyChecking` for convenience. This reduces protection against man-in-the-middle attacks. In production environments or untrusted networks, consider enabling host key verification.

✅ **Local Access**: The web dashboard and all tunnels are only accessible on localhost by default.

✅ **Service Authentication**: This tool provides tunnel access only. Ensure downstream services (Grafana, Prometheus, etc.) have proper authentication enabled.

### Threat Model

For detailed information about:
- What attacks are in scope
- What attacks are out of scope
- Security assumptions and boundaries

See [THREAT_MODEL.md](THREAT_MODEL.md).

### Reporting Security Issues

**Please do not report security vulnerabilities through public GitHub issues.**

See [SECURITY.md](SECURITY.md) for our security policy and reporting guidelines.

## Troubleshooting

### No ports found

- Verify SSH access to the server
- Check that the port range includes the services you're looking for
- Ensure `ss` or `netstat` is available on the remote server

### Tunnel creation fails

- Verify SSH key permissions (should be 600)
- Check that local ports are not already in use
- Ensure SSH access is working: `ssh user@server`

### Services not detected

- Services may not be HTTP/HTTPS based
- Check firewall rules on the remote server
- Verify services are actually running on the detected ports

## Performance

The tool is designed for efficiency. Key performance characteristics:

- **Port Scanning**: Parses port lists in microseconds
- **Service Detection**: HTTP probes with configurable timeouts (default 3s)
- **SSH Config Parsing**: Fast parsing of SSH configuration files
- **Memory Usage**: Minimal memory footprint, suitable for long-running sessions

### Benchmarks

Run benchmarks with:
```bash
go test -bench=. -benchmem ./pkg/...
```

Example results (Apple M4):
```
BenchmarkParsePortRange-10           15662553    77.46 ns/op
BenchmarkDeduplicatePorts-10          13049281    91.92 ns/op
BenchmarkParseSSOutput-10              892759  1383 ns/op
BenchmarkParseSSHConfig-10           1642608   715.6 ns/op
BenchmarkDetectServices-10            1817655   660.3 ns/op
```

## License

MIT

