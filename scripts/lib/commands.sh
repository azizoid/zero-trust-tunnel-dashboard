#!/bin/bash

is_valid_command() {
    case "$1" in
        build|test|bench|lint|vulncheck|clean|install|release|help)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

get_command_desc() {
    case "$1" in
        build) echo "Build the binary" ;;
        test) echo "Run tests" ;;
        bench) echo "Run benchmarks" ;;
        lint) echo "Run linter" ;;
        vulncheck) echo "Check for vulnerabilities" ;;
        clean) echo "Clean build artifacts" ;;
        install) echo "Install to GOPATH/bin" ;;
        release) echo "Build release binaries with checksums" ;;
        help) echo "Show this help message" ;;
    esac
}

show_commands() {
    echo "Available commands:"
    echo ""
    for cmd in build test bench lint vulncheck clean install release help; do
        printf "  %-15s %s\n" "$cmd" "$(get_command_desc "$cmd")"
    done
    echo ""
    echo "Examples:"
    echo "  ./run build                    # Build the binary"
    echo "  ./run test                     # Run tests"
    echo "  ./run --host my-server         # Run tunnel-dash with SSH host"
    echo "  ./run --server example.com --user admin  # Run with direct connection"
    echo ""
    echo "For tunnel-dash options, pass them directly:"
    echo "  ./run --host my-server --scan-ports 3000-9000"
}

