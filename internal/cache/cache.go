// Package cache handles caching of IANA RDAP bootstrap data
package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	// IANABootstrapURL is the official IANA RDAP bootstrap URL
	IANABootstrapURL = "https://data.iana.org/rdap/dns.json"

	// CacheTTL is how long the cache is considered valid
	CacheTTL = 24 * time.Hour

	// CacheDir is the directory name for cache files
	CacheDir = ".domaindetails"

	// BootstrapFile is the cached bootstrap filename
	BootstrapFile = "rdap-bootstrap.json"

	// MetaFile stores cache metadata
	MetaFile = "cache-meta.json"
)

// IANABootstrap represents the IANA RDAP bootstrap file structure
type IANABootstrap struct {
	Description string       `json:"description"`
	Publication string       `json:"publication"`
	Services    [][][]string `json:"services"`
	Version     string       `json:"version"`
}

// CacheMeta stores metadata about the cache
type CacheMeta struct {
	LastUpdated time.Time `json:"lastUpdated"`
	Version     string    `json:"version"`
	TLDCount    int       `json:"tldCount"`
}

// CacheInfo provides information about the cache
type CacheInfo struct {
	Path        string
	LastUpdated time.Time
	TLDCount    int
	Age         time.Duration
	IsValid     bool
}

// Cache manages the local RDAP bootstrap cache
type Cache struct {
	cacheDir string
}

// NewCache creates a new Cache instance
func NewCache() *Cache {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return &Cache{
		cacheDir: filepath.Join(homeDir, CacheDir),
	}
}

// GetRDAPServer returns the RDAP server URL for a given TLD
func (c *Cache) GetRDAPServer(tld string) (string, error) {
	bootstrap, err := c.getBootstrap()
	if err != nil {
		return "", err
	}

	// Search for the TLD in services
	for _, service := range bootstrap.Services {
		if len(service) < 2 {
			continue
		}
		tlds := service[0]
		urls := service[1]

		for _, t := range tlds {
			if t == tld && len(urls) > 0 {
				return urls[0], nil
			}
		}
	}

	return "", fmt.Errorf("no RDAP server found for TLD: %s", tld)
}

// getBootstrap returns the cached bootstrap data, fetching if needed
func (c *Cache) getBootstrap() (*IANABootstrap, error) {
	// Check if cache exists and is valid
	meta, err := c.getMeta()
	if err == nil && time.Since(meta.LastUpdated) < CacheTTL {
		// Cache is valid, read from file
		data, err := c.readBootstrap()
		if err == nil {
			return data, nil
		}
	}

	// Cache is invalid or missing, fetch new data
	if err := c.Update(); err != nil {
		// If update fails but we have stale cache, use it
		data, readErr := c.readBootstrap()
		if readErr == nil {
			return data, nil
		}
		return nil, fmt.Errorf("failed to fetch bootstrap data: %v", err)
	}

	return c.readBootstrap()
}

// readBootstrap reads the cached bootstrap file
func (c *Cache) readBootstrap() (*IANABootstrap, error) {
	path := filepath.Join(c.cacheDir, BootstrapFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var bootstrap IANABootstrap
	if err := json.Unmarshal(data, &bootstrap); err != nil {
		return nil, err
	}

	return &bootstrap, nil
}

// getMeta reads the cache metadata
func (c *Cache) getMeta() (*CacheMeta, error) {
	path := filepath.Join(c.cacheDir, MetaFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var meta CacheMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// Update fetches fresh bootstrap data from IANA
func (c *Cache) Update() error {
	// Ensure cache directory exists
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Fetch from IANA
	resp, err := http.Get(IANABootstrapURL)
	if err != nil {
		return fmt.Errorf("failed to fetch bootstrap data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("IANA returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Parse to validate and count TLDs
	var bootstrap IANABootstrap
	if err := json.Unmarshal(body, &bootstrap); err != nil {
		return fmt.Errorf("invalid bootstrap data: %v", err)
	}

	// Count TLDs
	tldCount := 0
	for _, service := range bootstrap.Services {
		if len(service) > 0 {
			tldCount += len(service[0])
		}
	}

	// Write bootstrap file
	bootstrapPath := filepath.Join(c.cacheDir, BootstrapFile)
	if err := os.WriteFile(bootstrapPath, body, 0644); err != nil {
		return fmt.Errorf("failed to write bootstrap file: %v", err)
	}

	// Write metadata
	meta := CacheMeta{
		LastUpdated: time.Now(),
		Version:     bootstrap.Version,
		TLDCount:    tldCount,
	}

	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	metaPath := filepath.Join(c.cacheDir, MetaFile)
	if err := os.WriteFile(metaPath, metaData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %v", err)
	}

	return nil
}

// Info returns information about the cache
func (c *Cache) Info() (*CacheInfo, error) {
	meta, err := c.getMeta()
	if err != nil {
		return nil, fmt.Errorf("cache not found or invalid: %v", err)
	}

	age := time.Since(meta.LastUpdated)

	return &CacheInfo{
		Path:        c.cacheDir,
		LastUpdated: meta.LastUpdated,
		TLDCount:    meta.TLDCount,
		Age:         age,
		IsValid:     age < CacheTTL,
	}, nil
}

// Clear removes all cached data
func (c *Cache) Clear() error {
	bootstrapPath := filepath.Join(c.cacheDir, BootstrapFile)
	metaPath := filepath.Join(c.cacheDir, MetaFile)

	os.Remove(bootstrapPath)
	os.Remove(metaPath)

	return nil
}
