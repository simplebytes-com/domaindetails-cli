#!/bin/bash
# Setup script for domaindetails CLI
# Run this to set up everything for Docker Hub and Homebrew

set -e

echo "=== domaindetails CLI Setup ==="
echo ""

# Check prerequisites
command -v go >/dev/null 2>&1 || { echo "Installing Go..."; brew install go; }
command -v gh >/dev/null 2>&1 || { echo "Installing GitHub CLI..."; brew install gh; }
command -v docker >/dev/null 2>&1 || { echo "Error: Docker not installed. Please install Docker Desktop."; exit 1; }

cd "$(dirname "$0")"

echo "1. Building locally..."
go mod tidy
go build -o bin/domaindetails ./cmd/domaindetails
echo "   ✓ Build successful"

echo ""
echo "2. Testing..."
./bin/domaindetails --version
./bin/domaindetails lookup example.com --json | head -5
echo "   ✓ Tests passed"

echo ""
echo "3. Switching to simplebytes-agent GitHub account..."
gh auth switch -u simplebytes-agent 2>/dev/null || echo "   (already on correct account or switch manually)"

echo ""
echo "4. Creating GitHub repository..."
if gh repo view simplebytes-com/domaindetails-cli >/dev/null 2>&1; then
    echo "   Repository already exists"
else
    gh repo create simplebytes-com/domaindetails-cli \
        --public \
        --description "Domain RDAP and WHOIS lookup CLI tool" \
        --homepage "https://domaindetails.com"
    echo "   ✓ Repository created"
fi

echo ""
echo "5. Creating Homebrew tap repository..."
if gh repo view simplebytes-com/homebrew-tap >/dev/null 2>&1; then
    echo "   Repository already exists"
else
    gh repo create simplebytes-com/homebrew-tap \
        --public \
        --description "Homebrew formulae for Simple Bytes tools"
    echo "   ✓ Homebrew tap created"

    # Set up the tap
    TMPDIR=$(mktemp -d)
    cd "$TMPDIR"
    git clone https://github.com/simplebytes-com/homebrew-tap.git
    cd homebrew-tap
    mkdir -p Formula
    cat > Formula/domaindetails.rb << 'FORMULA'
class Domaindetails < Formula
  desc "Domain RDAP and WHOIS lookup CLI tool"
  homepage "https://domaindetails.com"
  version "0.0.1"
  license "MIT"
  # Formula will be auto-updated on release
end
FORMULA
    cat > README.md << 'README'
# Homebrew Tap for Simple Bytes

This repository contains Homebrew formulae for tools built by [Simple Bytes LLC](https://domaindetails.com).

## Installation

```bash
brew tap simplebytes-com/tap
brew install domaindetails
```

## Available Formulae

- **domaindetails** - Domain RDAP and WHOIS lookup CLI tool
README
    git add .
    git commit -m "Initial commit"
    git push
    cd -
    rm -rf "$TMPDIR"
fi

echo ""
echo "6. Initializing git and pushing..."
cd "$(dirname "$0")"
if [ ! -d .git ]; then
    git init
    git add .
    git commit -m "Initial commit: domaindetails CLI

A fast CLI tool for domain RDAP and WHOIS lookups.

Features:
- RDAP lookups with local IANA bootstrap caching
- WHOIS lookups via DomainDetails.com API
- JSON and human-readable output formats
- Cross-platform (macOS, Linux, Windows)

Built by DomainDetails.com"
    git branch -M main
    git remote add origin https://github.com/simplebytes-com/domaindetails-cli.git
fi
git push -u origin main
echo "   ✓ Code pushed"

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Next steps:"
echo ""
echo "1. Set up Docker Hub:"
echo "   - Go to https://hub.docker.com"
echo "   - Create repository: domaindetails/cli"
echo "   - Create access token at Account Settings → Security"
echo ""
echo "2. Add GitHub Secrets at:"
echo "   https://github.com/simplebytes-com/domaindetails-cli/settings/secrets/actions"
echo ""
echo "   DOCKERHUB_USERNAME = your Docker Hub username"
echo "   DOCKERHUB_TOKEN    = your Docker Hub access token"
echo "   HOMEBREW_TAP_TOKEN = GitHub PAT with repo scope"
echo ""
echo "3. Create first release:"
echo "   git tag v1.0.0"
echo "   git push origin v1.0.0"
echo ""
echo "This will automatically:"
echo "   - Build binaries for all platforms"
echo "   - Create GitHub release"
echo "   - Push Docker images"
echo "   - Update Homebrew formula"
