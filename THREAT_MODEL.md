# Threat Model

## Overview

This document describes the threat model for the Zero-Trust Tunnel Dashboard. It identifies what attacks are in scope, what is explicitly out of scope, and the security assumptions we make.

## Security Assumptions

1. **SSH Infrastructure is Trusted**: We assume the SSH server and key management are secure
2. **Local Machine is Secure**: The tool runs on a trusted local machine
3. **Remote Server Access**: The user has legitimate SSH access to the remote server
4. **Network Isolation**: The local dashboard is not exposed to untrusted networks

## Attacks In Scope

### 1. SSH Tunnel Security

**Threat**: Man-in-the-middle attacks on SSH connections

**Mitigation**: 
- SSH protocol provides encryption and authentication
- SSH host key verification (can be enabled, currently disabled for convenience)
- Private key authentication required

**Status**: Partially mitigated (host key checking disabled by default)

### 2. Local Port Exposure

**Threat**: Unauthorized access to local tunnel ports

**Mitigation**:
- Tunnels bind to localhost only (127.0.0.1)
- Dashboard serves on localhost only
- No network exposure by default

**Status**: ✅ Mitigated

### 3. Service Detection Probes

**Threat**: Service detection probes may be logged or trigger alerts

**Mitigation**:
- Probes are minimal HTTP requests
- Only to services already accessible via SSH
- User is aware of what services are being accessed

**Status**: ✅ Acceptable risk (by design)

### 4. SSH Key Compromise

**Threat**: If SSH private key is compromised, attacker gains access

**Mitigation**:
- Key permissions should be 600
- Keys should be stored securely
- Use SSH agent forwarding when possible
- Rotate keys regularly

**Status**: ✅ User responsibility (documented)

### 5. Remote Service Vulnerabilities

**Threat**: Vulnerable services accessible through tunnels

**Mitigation**:
- This tool only provides access, doesn't fix service vulnerabilities
- Users should ensure services are properly secured
- Services should have their own authentication

**Status**: ⚠️ Out of scope (service-level security)

## Attacks Out of Scope

### 1. Compromised Remote Server

**Not In Scope**: If the remote server is already compromised, this tool cannot protect against that. The tool assumes the remote server is trusted.

**Rationale**: This is a server security issue, not a tool security issue.

### 2. Vulnerable Downstream Services

**Not In Scope**: If Grafana, Prometheus, or other services have vulnerabilities, this tool does not mitigate those.

**Rationale**: Service security is the responsibility of service maintainers and operators.

### 3. SSH Server Vulnerabilities

**Not In Scope**: Vulnerabilities in the SSH server software itself (OpenSSH, etc.)

**Rationale**: We rely on SSH as a trusted component. Users should keep SSH servers updated.

### 4. Local Machine Compromise

**Not In Scope**: If the local machine running this tool is compromised, all security guarantees are void.

**Rationale**: This is a fundamental security assumption. If the local machine is compromised, the attacker already has access.

### 5. Network-Level Attacks on SSH

**Not In Scope**: Attacks that compromise the SSH protocol itself or the underlying network infrastructure.

**Rationale**: We assume SSH and the network layer are secure. This is standard for SSH-based tools.

### 6. Social Engineering

**Not In Scope**: Attacks that trick users into running malicious commands or exposing credentials.

**Rationale**: User education and operational security are out of scope for this tool.

## Security Boundaries

### What This Tool Protects

✅ **Network Isolation**: Services are not exposed to the internet  
✅ **Encrypted Transport**: All traffic goes through SSH tunnels  
✅ **Local Access Control**: Only localhost access by default  
✅ **No Direct Network Exposure**: Remote services remain behind firewall  

### What This Tool Does NOT Protect

❌ **Service Authentication**: Services must implement their own auth  
❌ **Service Vulnerabilities**: Does not fix bugs in downstream services  
❌ **SSH Key Security**: Relies on user to protect SSH keys  
❌ **Compromised Servers**: Cannot protect against already-compromised servers  
❌ **Malicious Services**: Does not protect against malicious services on remote server  

## Zero-Trust Model

This tool implements a **zero-trust tunnel** approach:

1. **No Trust in Network**: We assume the network is untrusted
2. **Encrypted Tunnels**: All communication is encrypted via SSH
3. **No Direct Exposure**: Services are never directly exposed
4. **Principle of Least Privilege**: Only necessary ports are tunneled
5. **Local Access Only**: Services are accessed through localhost

### Zero-Trust Principles Applied

- ✅ **Verify Explicitly**: SSH authentication required
- ✅ **Use Least Privilege**: Only tunnel necessary ports
- ✅ **Assume Breach**: Services remain behind firewall, not exposed

### Zero-Trust Limitations

- ⚠️ **SSH Host Key Verification**: Currently disabled (convenience vs security trade-off)
- ⚠️ **Service Authentication**: Not enforced by this tool
- ⚠️ **Audit Logging**: Limited logging of access patterns

## Recommendations for Production Use

1. **Enable SSH Host Key Checking**: Modify the tool or SSH config to enable host key verification
2. **Use SSH Certificates**: Consider using SSH certificates instead of keys for better key management
3. **Monitor Tunnel Activity**: Log and monitor tunnel creation and usage
4. **Restrict Dashboard Access**: Ensure dashboard port is firewall-protected
5. **Regular Key Rotation**: Rotate SSH keys regularly
6. **Service Authentication**: Ensure all services have proper authentication enabled
7. **Network Segmentation**: Use network segmentation to limit blast radius

## Threat Matrix

| Threat | Likelihood | Impact | Mitigation | Status |
|--------|-----------|--------|------------|--------|
| MITM on SSH | Medium | High | SSH encryption, host key verification (optional) | ⚠️ Partially mitigated |
| Local port exposure | Low | Medium | localhost-only binding | ✅ Mitigated |
| SSH key compromise | Low | Critical | Key management best practices | ✅ User responsibility |
| Service vulnerabilities | Medium | High | Service-level security | ❌ Out of scope |
| Compromised remote server | Low | Critical | Server security | ❌ Out of scope |

## Updates

This threat model will be updated as:
- New threats are identified
- The tool's architecture changes
- Security best practices evolve

Last updated: 2024

