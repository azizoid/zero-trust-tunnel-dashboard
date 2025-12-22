# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Version information support (`--version` flag)
- Security policy documentation (SECURITY.md)
- Threat model documentation (THREAT_MODEL.md)
- CHANGELOG.md for tracking changes

### Changed
- Improved security documentation in README

## [0.1.0] - 2024-XX-XX

### Added
- Initial release
- SSH tunnel management for remote services
- Automatic port scanning via SSH
- Service detection (Grafana, Prometheus, Kubernetes, Jenkins, Jupyter, etc.)
- Web dashboard for accessing services
- CLI interface with service information
- Docker container detection support
- Nginx Proxy Manager integration
- Support for SSH config file (`~/.ssh/config`)
- Multiple detection modes (docker, direct, both)

### Security
- SSH tunnel-based zero-trust access model
- Localhost-only binding for tunnels and dashboard
- SSH key authentication support

[Unreleased]: https://github.com/azizoid/zero-trust-tunnel-dashboard/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/azizoid/zero-trust-tunnel-dashboard/releases/tag/v0.1.0


