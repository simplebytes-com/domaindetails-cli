// Package types defines common types used across the CLI
package types

// LookupResult represents the result of a domain lookup
type LookupResult struct {
	Domain    string      `json:"domain"`
	Available bool        `json:"available"`
	Method    string      `json:"method"`
	Message   string      `json:"message,omitempty"`
	Parsed    *ParsedData `json:"parsed,omitempty"`
	Raw       string      `json:"raw,omitempty"`
}

// ParsedData contains parsed domain registration information
type ParsedData struct {
	DomainName     string   `json:"domainName,omitempty"`
	Registrar      string   `json:"registrar,omitempty"`
	Registrant     string   `json:"registrant,omitempty"`
	CreationDate   string   `json:"creationDate,omitempty"`
	ExpirationDate string   `json:"expirationDate,omitempty"`
	LastModified   string   `json:"lastModified,omitempty"`
	Nameservers    []string `json:"nameservers,omitempty"`
	Status         []string `json:"status,omitempty"`
	DNSSEC         string   `json:"dnssec,omitempty"`
	WhoisServer    string   `json:"whoisServer,omitempty"`
}
