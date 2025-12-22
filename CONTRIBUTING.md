# Contributing to Zero-Trust Tunnel Dashboard

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Basic understanding of SSH and networking concepts

### Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/your-username/zero-trust-tunnel-dashboard.git
   cd zero-trust-tunnel-dashboard
   ```

2. **Verify the setup**
   ```bash
   go version  # Should be 1.21+
   go test ./...
   ```

3. **Build the project**
   ```bash
   go build -o tunnel-dash ./cmd/tunnel-dash
   ```

## Development Workflow

### Making Changes

1. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes**
   - Write clear, readable code
   - Add tests for new functionality
   - Update documentation as needed

3. **Run tests and linters**
   ```bash
   # Run all tests
   go test ./...
   
   # Run tests with race detection
   go test -race ./...
   
   # Run benchmarks
   go test -bench=. -benchmem ./...
   
   # Run linter (if installed locally)
   golangci-lint run
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `chore:` - Maintenance tasks

Examples:
```
feat: add support for custom port ranges
fix: handle SSH connection timeout gracefully
docs: update README with new examples
test: add benchmarks for service detection
```

### Pull Request Process

1. **Push your branch**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request**
   - Use a clear, descriptive title
   - Describe what changes you made and why
   - Reference any related issues
   - Ensure all CI checks pass

3. **Address feedback**
   - Respond to review comments
   - Make requested changes
   - Update your PR as needed

## Code Style

### Go Style Guide

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` to format code
- Follow the existing code style in the project
- Keep functions focused and small
- Write clear, self-documenting code

### Linting

We use `golangci-lint` for code quality. The configuration is in `.golangci.yml`.

Common issues to avoid:
- Unused variables or imports
- Missing error handling
- Inefficient code patterns
- Missing comments on exported functions

### Testing

- Write tests for new functionality
- Aim for good test coverage
- Use table-driven tests when appropriate
- Test error cases, not just happy paths
- Add benchmarks for performance-critical code

Example test structure:
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "result",
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Project Structure

```
zero-trust-tunnel-dashboard/
├── cmd/
│   └── tunnel-dash/      # Main application entry point
├── pkg/
│   ├── config/           # Configuration management
│   ├── dashboard/        # Dashboard generation
│   ├── detector/         # Service detection
│   ├── scanner/          # Port scanning
│   ├── server/           # HTTP server
│   ├── ssh/              # SSH client utilities
│   ├── sshconfig/        # SSH config parsing
│   ├── tunnel/           # Tunnel management
│   └── version/          # Version information
├── .github/
│   └── workflows/        # CI/CD workflows
├── README.md
├── CONTRIBUTING.md
├── SECURITY.md
└── THREAT_MODEL.md
```

## Areas for Contribution

### Good First Issues

Look for issues labeled `good first issue` - these are great for newcomers!

### Areas That Need Help

- **Performance improvements**: Benchmarks and optimizations
- **Documentation**: Examples, tutorials, architecture diagrams
- **Testing**: Increase test coverage
- **Features**: See open issues for feature requests
- **Bug fixes**: Check the issues list

## Security Contributions

If you find a security vulnerability, **please do not open a public issue**. Instead, see [SECURITY.md](SECURITY.md) for reporting guidelines.

## Questions?

- Open a GitHub Discussion for questions
- Check existing issues for similar questions
- Review the code and documentation

## Recognition

Contributors will be recognized in:
- Release notes
- Project documentation
- GitHub contributors list

Thank you for contributing! 🎉

