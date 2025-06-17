#!/bin/bash

# Get the version from git tag, or use a default if no tags exist
VERSION=$(git describe --tags 2>/dev/null || echo "dev")

# Get the current commit hash
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Get the build time
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')

echo "Building vizb with:"
echo "  Version:    $VERSION"
echo "  Commit:     $COMMIT"
echo "  Build Time: $BUILD_TIME"
echo

# Build the binary with ldflags to inject version information
go build -ldflags="-X 'github.com/goptics/vizb/cmd.Version=$VERSION' \
                   -X 'github.com/goptics/vizb/cmd.CommitSHA=$COMMIT' \
                   -X 'github.com/goptics/vizb/cmd.BuildTime=$BUILD_TIME'"

echo "Build complete!"
