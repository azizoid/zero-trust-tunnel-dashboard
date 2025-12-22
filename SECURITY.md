# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

### Go Version Requirements

- **Minimum**: Go 1.21
- **Recommended**: Go 1.24+ (includes fixes for standard library vulnerabilities)

Note: Go 1.23.x contains known vulnerabilities in the standard library (crypto/x509, net/http, etc.) that are fixed in Go 1.24+. We recommend upgrading to Go 1.24+ for production use.

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **Email**: Send details to the repository maintainer (see GitHub profile)
2. **GitHub Security Advisory**: Use the [GitHub Security Advisory](https://github.com/azizoid/zero-trust-tunnel-dashboard/security/advisories/new) feature

### What to Include

When reporting a vulnerability, please include:

- **Type of vulnerability** (e.g., authentication bypass, information disclosure, etc.)
- **Affected component** (e.g., tunnel manager, SSH client, service detector)
- **Steps to reproduce** (if applicable)
- **Potential impact** (what an attacker could achieve)
- **Suggested fix** (if you have one)

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution**: Depends on severity and complexity

### Security Best Practices

This tool is designed for zero-trust access to remote services. However, users should be aware of:

1. **SSH Key Security**: Protect your SSH private keys with appropriate permissions (600)
2. **Host Key Verification**: The tool disables SSH host key checking for convenience. In production, consider enabling it
3. **Local Dashboard**: The web dashboard is only accessible on localhost by default
4. **Network Exposure**: Ensure the dashboard port is not exposed to untrusted networks
5. **Service Authentication**: This tool provides tunnel access only. Ensure downstream services have proper authentication

### Known Security Considerations

- **SSH Host Key Checking Disabled**: The tool disables `StrictHostKeyChecking` for convenience. This reduces protection against man-in-the-middle attacks. Use with caution in untrusted networks.
- **Local Port Binding**: Tunnels bind to localhost only, but ensure no unauthorized access to your local machine
- **Service Detection**: The tool probes HTTP/HTTPS endpoints. Some services may log these probes

### Security Model

This tool implements a **zero-trust tunnel** model:

- **No Trust in Network**: All communication goes through SSH tunnels
- **No Direct Exposure**: Remote services are not exposed to the internet
- **Local Access Only**: Services are accessed through localhost tunnels
- **SSH Authentication**: Access requires valid SSH credentials

See [THREAT_MODEL.md](THREAT_MODEL.md) for detailed threat analysis.

### Responsible Disclosure

We follow responsible disclosure practices:

1. Reporter notifies maintainers privately
2. Maintainers confirm and assess the vulnerability
3. A fix is developed and tested
4. A security advisory is published
5. The fix is released

We appreciate your help in keeping this project secure.

