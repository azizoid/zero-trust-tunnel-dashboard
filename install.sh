#!/bin/bash
# One-command install for tunnel-dash
# Usage: curl -sSL https://raw.githubusercontent.com/azizoid/zero-trust-tunnel-dashboard/main/install.sh | bash

set -e

echo "Installing tunnel-dash..."
go install github.com/azizoid/zero-trust-tunnel-dashboard@latest

if [ $? -eq 0 ]; then
    echo "✓ tunnel-dash installed successfully!"
    echo "Run 'tunnel-dash --help' to get started"
else
    echo "✗ Installation failed"
    exit 1
fi
