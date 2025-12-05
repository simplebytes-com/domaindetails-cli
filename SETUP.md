# Setup Guide for domaindetails CLI

This guide walks through the complete setup process to publish the CLI to Docker Hub and Homebrew.

## Prerequisites

- Go 1.21+ (`brew install go`)
- Docker Desktop
- GitHub CLI (`brew install gh`)
- GitHub account with access to `simplebytes-com` organization

## Step 1: Build and Test Locally

```bash
cd /Users/julianengeltcg/projects_code/domaindetails/domaindetails-cli

# Download dependencies
go mod tidy

# Build
go build -o bin/domaindetails ./cmd/domaindetails

# Test
./bin/domaindetails --help
./bin/domaindetails lookup example.com
./bin/domaindetails rdap google.com --json
```

## Step 2: Create GitHub Repository

```bash
# Switch to SimpleBytes Agent account
gh auth switch
# Select: simplebytes-agent

# Create the repository
gh repo create simplebytes-com/domaindetails-cli \
  --public \
  --description "Domain RDAP and WHOIS lookup CLI tool" \
  --homepage "https://domaindetails.com"

# Initialize git and push
cd /Users/julianengeltcg/projects_code/domaindetails/domaindetails-cli
git init
git add .
git commit -m "Initial commit: domaindetails CLI"
git branch -M main
git remote add origin https://github.com/simplebytes-com/domaindetails-cli.git
git push -u origin main
```

## Step 3: Create Homebrew Tap Repository

```bash
# Create the homebrew tap repo
gh repo create simplebytes-com/homebrew-tap \
  --public \
  --description "Homebrew formulae for Simple Bytes tools"

# Clone and set up
cd /tmp
git clone https://github.com/simplebytes-com/homebrew-tap.git
cd homebrew-tap
mkdir -p Formula

# Create initial formula (will be auto-updated on release)
cat > Formula/domaindetails.rb << 'EOF'
class Domaindetails < Formula
  desc "Domain RDAP and WHOIS lookup CLI tool"
  homepage "https://domaindetails.com"
  version "0.0.1"
  license "MIT"

  on_macos do
    url "https://github.com/simplebytes-com/domaindetails-cli/releases/download/v0.0.1/domaindetails-0.0.1-darwin-arm64.tar.gz"
    sha256 "placeholder"
  end

  def install
    bin.install "domaindetails"
  end

  test do
    system "#{bin}/domaindetails", "--version"
  end
end
EOF

git add .
git commit -m "Initial commit: homebrew tap"
git push -u origin main
```

## Step 4: Set Up Docker Hub

1. Go to https://hub.docker.com
2. Create organization `domaindetails` (if not exists)
3. Create repository `domaindetails/cli`
4. Go to Account Settings → Security → Access Tokens
5. Create a new access token with Read/Write permissions
6. Save the token for GitHub secrets

## Step 5: Configure GitHub Secrets

Go to https://github.com/simplebytes-com/domaindetails-cli/settings/secrets/actions

Add these secrets:

| Secret Name | Description |
|-------------|-------------|
| `DOCKERHUB_USERNAME` | Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub access token |
| `HOMEBREW_TAP_TOKEN` | GitHub PAT with repo access to homebrew-tap |

### Creating HOMEBREW_TAP_TOKEN

1. Go to https://github.com/settings/tokens
2. Generate new token (classic)
3. Select scopes: `repo` (full control)
4. Copy token and add as secret

## Step 6: Create First Release

```bash
cd /Users/julianengeltcg/projects_code/domaindetails/domaindetails-cli

# Tag the release
git tag v1.0.0
git push origin v1.0.0
```

This will trigger the GitHub Actions workflow which will:
1. Build binaries for all platforms
2. Create a GitHub release with the binaries
3. Build and push Docker images to Docker Hub
4. Update the Homebrew formula automatically

## Step 7: Verify Installation Methods

After the release workflow completes:

### Docker
```bash
docker pull domaindetails/cli:1.0.0
docker run domaindetails/cli lookup example.com
```

### Homebrew
```bash
brew tap simplebytes-com/tap
brew install domaindetails
domaindetails lookup example.com
```

## Backlinks

The following pages will link back to domaindetails.com:

1. **Docker Hub**: https://hub.docker.com/r/domaindetails/cli
   - Repository description
   - README with link to website

2. **Homebrew**: https://github.com/simplebytes-com/homebrew-tap
   - Formula homepage field
   - README

3. **GitHub**: https://github.com/simplebytes-com/domaindetails-cli
   - Repository homepage
   - README
   - Release notes

## Troubleshooting

### Go not found
```bash
brew install go
export PATH=$PATH:/opt/homebrew/bin
```

### Docker build fails
```bash
# Make sure Docker Desktop is running
docker info

# Build manually
docker build -t domaindetails/cli:latest .
```

### GitHub Actions workflow fails
- Check the Actions tab for error logs
- Ensure all secrets are configured correctly
- Verify repository permissions
