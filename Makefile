# domaindetails CLI Makefile

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -w -s"

BINARY := domaindetails
BUILD_DIR := bin
DIST_DIR := dist

# Go settings
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

.PHONY: all build clean test install docker homebrew

all: build

## Build

build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/domaindetails

build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY)-linux-amd64 ./cmd/domaindetails
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY)-linux-arm64 ./cmd/domaindetails

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY)-darwin-amd64 ./cmd/domaindetails
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY)-darwin-arm64 ./cmd/domaindetails

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY)-windows-amd64.exe ./cmd/domaindetails

## Install

install: build
	@echo "Installing $(BINARY)..."
	cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)

## Test

test:
	$(GOTEST) -v ./...

## Clean

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)

## Dependencies

deps:
	$(GOMOD) download
	$(GOMOD) tidy

## Docker

docker:
	@echo "Building Docker image..."
	docker build -t domaindetails/cli:$(VERSION) -t domaindetails/cli:latest .

docker-push:
	@echo "Pushing Docker image..."
	docker push domaindetails/cli:$(VERSION)
	docker push domaindetails/cli:latest

## Release (creates archives for GitHub releases)

release: clean build-all
	@echo "Creating release archives..."
	@mkdir -p $(DIST_DIR)/release
	cd $(DIST_DIR) && tar -czf release/$(BINARY)-$(VERSION)-linux-amd64.tar.gz $(BINARY)-linux-amd64
	cd $(DIST_DIR) && tar -czf release/$(BINARY)-$(VERSION)-linux-arm64.tar.gz $(BINARY)-linux-arm64
	cd $(DIST_DIR) && tar -czf release/$(BINARY)-$(VERSION)-darwin-amd64.tar.gz $(BINARY)-darwin-amd64
	cd $(DIST_DIR) && tar -czf release/$(BINARY)-$(VERSION)-darwin-arm64.tar.gz $(BINARY)-darwin-arm64
	cd $(DIST_DIR) && zip release/$(BINARY)-$(VERSION)-windows-amd64.zip $(BINARY)-windows-amd64.exe
	@echo "Release archives created in $(DIST_DIR)/release/"

## Help

help:
	@echo "domaindetails CLI build targets:"
	@echo ""
	@echo "  build        - Build for current platform"
	@echo "  build-all    - Build for all platforms"
	@echo "  install      - Install to /usr/local/bin"
	@echo "  test         - Run tests"
	@echo "  clean        - Remove build artifacts"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  docker       - Build Docker image"
	@echo "  docker-push  - Push Docker image to registry"
	@echo "  release      - Create release archives"
	@echo "  help         - Show this help"
