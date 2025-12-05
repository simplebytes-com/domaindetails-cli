// Package whois provides WHOIS lookup functionality via the DomainDetails.com API
package whois

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/simplebytes-com/domaindetails-cli/internal/types"
)

const (
	// APIBaseURL is the DomainDetails.com API endpoint
	APIBaseURL = "https://api.domaindetails.io"

	// RequestTimeout is the timeout for WHOIS requests
	RequestTimeout = 15 * time.Second

	// UserAgent identifies the CLI to the API
	UserAgent = "domaindetails-cli/1.0 (https://domaindetails.com)"
)

// Client performs WHOIS lookups via the DomainDetails.com API
type Client struct {
	verbose bool
	client  *http.Client
	baseURL string
}

// APIResponse represents the response from the WHOIS API
type APIResponse struct {
	ParsedData *APIParsedData `json:"parsedData"`
	RawData    string         `json:"rawData"`
	Error      string         `json:"error,omitempty"`
}

// APIParsedData represents the parsed WHOIS data from the API
type APIParsedData struct {
	DomainName     string   `json:"domainName"`
	Registrar      string   `json:"registrar"`
	Registrant     string   `json:"registrant"`
	CreationDate   string   `json:"creationDate"`
	ExpirationDate string   `json:"expirationDate"`
	LastModified   string   `json:"lastModified"`
	Nameservers    []string `json:"nameservers"`
	Status         []string `json:"status"`
	DNSSEC         string   `json:"dnssec"`
	WhoisServer    string   `json:"whoisServer"`
}

// NewClient creates a new WHOIS client
func NewClient(verbose bool) *Client {
	return &Client{
		verbose: verbose,
		client: &http.Client{
			Timeout: RequestTimeout,
		},
		baseURL: APIBaseURL,
	}
}

// Lookup performs a WHOIS lookup for the given domain via the API
func (c *Client) Lookup(domain string) (*types.LookupResult, error) {
	// Build API URL
	apiURL := fmt.Sprintf("%s/api/whois?domain=%s", c.baseURL, url.QueryEscape(domain))

	if c.verbose {
		fmt.Printf("Querying WHOIS API: %s\n", apiURL)
	}

	// Make request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/json")
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
			Method:    "whois",
			Message:   "Domain not found",
		}, nil
	}

	if resp.StatusCode != 200 {
		var errResp struct {
			Error string `json:"error"`
		}
		json.Unmarshal(body, &errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("API error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if apiResp.Error != "" {
		return nil, fmt.Errorf("API error: %s", apiResp.Error)
	}

	// Convert to common result format
	result := c.convertToResult(domain, &apiResp)

	return result, nil
}

// convertToResult converts API response to common result format
func (c *Client) convertToResult(domain string, resp *APIResponse) *types.LookupResult {
	parsed := &types.ParsedData{}

	if resp.ParsedData != nil {
		parsed.DomainName = resp.ParsedData.DomainName
		parsed.Registrar = resp.ParsedData.Registrar
		parsed.Registrant = resp.ParsedData.Registrant
		parsed.CreationDate = resp.ParsedData.CreationDate
		parsed.ExpirationDate = resp.ParsedData.ExpirationDate
		parsed.LastModified = resp.ParsedData.LastModified
		parsed.Nameservers = resp.ParsedData.Nameservers
		parsed.Status = resp.ParsedData.Status
		parsed.DNSSEC = resp.ParsedData.DNSSEC
		parsed.WhoisServer = resp.ParsedData.WhoisServer
	}

	return &types.LookupResult{
		Domain:    domain,
		Available: false,
		Method:    "whois",
		Parsed:    parsed,
		Raw:       resp.RawData,
	}
}
