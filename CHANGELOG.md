# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] - 2025-12-23

### Changed
- Refactored dashboard generator for better maintainability and testability
- Separated view model logic from rendering
- Moved HTML template to embedded file
- Improved service access resolution logic
- Refactored Docker service detection with matcher table pattern
- Centralized port parsing logic
- Removed URL generation from detector (zero-trust principle)
- Improved error handling for SSH commands
- Simplified CI workflow to single Go version (1.23)
- Updated minimum Go version requirement to 1.23
- Improved CI caching and performance
- Fixed golangci-lint configuration for v2.7.2 compatibility

### Fixed
- Dashboard internal service counting logic
- Network name normalization for consistent grouping
- Unused parameter warnings in tests

## [1.1.1] - Previous release

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

[Unreleased]: https://github.com/azizoid/zero-trust-tunnel-dashboard/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/azizoid/zero-trust-tunnel-dashboard/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/azizoid/zero-trust-tunnel-dashboard/compare/v1.1.0...v1.1.1
[0.1.0]: https://github.com/azizoid/zero-trust-tunnel-dashboard/releases/tag/v0.1.0
