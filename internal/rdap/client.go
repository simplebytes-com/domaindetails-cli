// Package rdap provides RDAP lookup functionality
package rdap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/simplebytes-com/domaindetails-cli/internal/cache"
	"github.com/simplebytes-com/domaindetails-cli/internal/types"
)

const (
	// RequestTimeout is the timeout for RDAP requests
	RequestTimeout = 10 * time.Second

	// UserAgent identifies the CLI to RDAP servers
	UserAgent = "domaindetails-cli/1.0 (https://domaindetails.com)"
)

// Client performs RDAP lookups
type Client struct {
	cache   *cache.Cache
	verbose bool
	client  *http.Client
}

// NewClient creates a new RDAP client
func NewClient(verbose bool) *Client {
	return &Client{
		cache:   cache.NewCache(),
		verbose: verbose,
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// RDAPResponse represents the raw RDAP response
type RDAPResponse struct {
	ObjectClassName string            `json:"objectClassName"`
	LDHName         string            `json:"ldhName"`
	UnicodeName     string            `json:"unicodeName,omitempty"`
	Handle          string            `json:"handle,omitempty"`
	Status          []string          `json:"status,omitempty"`
	Events          []RDAPEvent       `json:"events,omitempty"`
	Entities        []RDAPEntity      `json:"entities,omitempty"`
	Nameservers     []RDAPNameserver  `json:"nameservers,omitempty"`
	SecureDNS       *RDAPSecureDNS    `json:"secureDNS,omitempty"`
	Links           []RDAPLink        `json:"links,omitempty"`
	Remarks         []RDAPRemark      `json:"remarks,omitempty"`
	Port43          string            `json:"port43,omitempty"`
	ErrorCode       int               `json:"errorCode,omitempty"`
	Title           string            `json:"title,omitempty"`
	Description     []string          `json:"description,omitempty"`
}

// RDAPEvent represents an RDAP event
type RDAPEvent struct {
	EventAction string `json:"eventAction"`
	EventDate   string `json:"eventDate"`
	EventActor  string `json:"eventActor,omitempty"`
}

// RDAPEntity represents an RDAP entity
type RDAPEntity struct {
	Handle     string          `json:"handle,omitempty"`
	Roles      []string        `json:"roles,omitempty"`
	VCardArray json.RawMessage `json:"vcardArray,omitempty"`
	Entities   []RDAPEntity    `json:"entities,omitempty"`
}

// RDAPNameserver represents an RDAP nameserver
type RDAPNameserver struct {
	LDHName     string `json:"ldhName"`
	UnicodeName string `json:"unicodeName,omitempty"`
}

// RDAPSecureDNS represents DNSSEC information
type RDAPSecureDNS struct {
	DelegationSigned bool `json:"delegationSigned"`
}

// RDAPLink represents an RDAP link
type RDAPLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
	Type string `json:"type,omitempty"`
}

// RDAPRemark represents an RDAP remark
type RDAPRemark struct {
	Title       string   `json:"title,omitempty"`
	Description []string `json:"description,omitempty"`
}

// Lookup performs an RDAP lookup for the given domain
func (c *Client) Lookup(domain string) (*types.LookupResult, error) {
	// Extract TLD
	tld := extractTLD(domain)

	if c.verbose {
		fmt.Printf("Extracting TLD: %s\n", tld)
	}

	// Get RDAP server from cache
	serverURL, err := c.cache.GetRDAPServer(tld)
	if err != nil {
		return nil, fmt.Errorf("no RDAP server for TLD .%s: %v", tld, err)
	}

	if c.verbose {
		fmt.Printf("Using RDAP server: %s\n", serverURL)
	}

	// Build query URL
	queryURL := fmt.Sprintf("%sdomain/%s", serverURL, domain)

	if c.verbose {
		fmt.Printf("Querying: %s\n", queryURL)
	}

	// Make request
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/rdap+json, application/json")
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Check for errors
	if resp.StatusCode == 404 {
		return &types.LookupResult{
			Domain:    domain,
			Available: true,
			Method:    "rdap",
			Message:   "Domain not found in registry",
		}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("RDAP server returned status %d", resp.StatusCode)
	}

	// Parse response
	var rdapResp RDAPResponse
	if err := json.Unmarshal(body, &rdapResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert to common result format
	result := c.convertToResult(domain, &rdapResp, body)

	return result, nil
}

// convertToResult converts RDAP response to common result format
func (c *Client) convertToResult(domain string, resp *RDAPResponse, raw []byte) *types.LookupResult {
	parsed := types.ParsedData{
		DomainName: resp.LDHName,
		Status:     resp.Status,
	}

	// Extract dates from events
	for _, event := range resp.Events {
		switch event.EventAction {
		case "registration":
			parsed.CreationDate = event.EventDate
		case "expiration":
			parsed.ExpirationDate = event.EventDate
		case "last changed":
			parsed.LastModified = event.EventDate
		}
	}

	// Extract nameservers
	for _, ns := range resp.Nameservers {
		parsed.Nameservers = append(parsed.Nameservers, ns.LDHName)
	}

	// Extract registrar
	for _, entity := range resp.Entities {
		for _, role := range entity.Roles {
			if role == "registrar" {
				parsed.Registrar = extractEntityName(entity)
			}
			if role == "registrant" {
				parsed.Registrant = extractEntityName(entity)
			}
		}
	}

	// DNSSEC
	if resp.SecureDNS != nil {
		if resp.SecureDNS.DelegationSigned {
			parsed.DNSSEC = "signed"
		} else {
			parsed.DNSSEC = "unsigned"
		}
	}

	return &types.LookupResult{
		Domain:    domain,
		Available: false,
		Method:    "rdap",
		Parsed:    &parsed,
		Raw:       string(raw),
	}
}

// extractTLD extracts the TLD from a domain
func extractTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return ""
	}

	// Check for common second-level TLDs
	secondLevel := map[string]bool{
		"co.uk": true, "org.uk": true, "me.uk": true,
		"com.au": true, "net.au": true, "org.au": true,
		"co.nz": true, "net.nz": true, "org.nz": true,
		"co.za": true, "co.in": true, "com.br": true,
	}

	if len(parts) >= 3 {
		twoLevel := parts[len(parts)-2] + "." + parts[len(parts)-1]
		if secondLevel[twoLevel] {
			return twoLevel
		}
	}

	return parts[len(parts)-1]
}

// extractEntityName tries to extract a name from an RDAP entity
func extractEntityName(entity RDAPEntity) string {
	if entity.Handle != "" {
		return entity.Handle
	}

	// Try to parse vCard if available
	if len(entity.VCardArray) > 0 {
		var vcard []interface{}
		if err := json.Unmarshal(entity.VCardArray, &vcard); err == nil {
			if len(vcard) >= 2 {
				if fields, ok := vcard[1].([]interface{}); ok {
					for _, field := range fields {
						if f, ok := field.([]interface{}); ok && len(f) >= 4 {
							if f[0] == "fn" {
								if name, ok := f[3].(string); ok {
									return name
								}
							}
						}
					}
				}
			}
		}
	}

	return ""
}
