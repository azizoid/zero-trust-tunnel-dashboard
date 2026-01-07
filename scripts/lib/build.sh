#!/bin/bash
# Build functions

# Function to check if build is needed
needs_build() {
    if [ ! -f "./tunnel-dash" ]; then
        return 0
    fi
    
    if [ -n "$(find ./cmd ./pkg -type f -name '*.go' -newer ./tunnel-dash 2>/dev/null | head -1)" ]; then
        return 0
    fi
    
    return 1
}

# Function to build tunnel-dash
build_tunnel_dash() {
    if needs_build; then
        echo "Building tunnel-dash..."
        make build > /dev/null 2>&1 || make build
    fi
}

