// Package output handles formatting and printing of lookup results
package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/simplebytes-com/domaindetails-cli/internal/types"
)

// Printer handles output formatting
type Printer struct {
	jsonOutput bool
	rawOutput  bool
}

// NewPrinter creates a new Printer
func NewPrinter(jsonOutput, rawOutput bool) *Printer {
	return &Printer{
		jsonOutput: jsonOutput,
		rawOutput:  rawOutput,
	}
}

// Print outputs the lookup result
func (p *Printer) Print(result *types.LookupResult) error {
	if p.jsonOutput {
		return p.printJSON(result)
	}
	return p.printText(result)
}

// printJSON outputs the result as JSON
func (p *Printer) printJSON(result *types.LookupResult) error {
	output := result

	// Remove raw data if not requested
	if !p.rawOutput {
		output = &types.LookupResult{
			Domain:    result.Domain,
			Available: result.Available,
			Method:    result.Method,
			Message:   result.Message,
			Parsed:    result.Parsed,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	fmt.Println(string(data))
	return nil
}

// printText outputs the result as formatted text
func (p *Printer) printText(result *types.LookupResult) error {
	fmt.Printf("\n%s\n", strings.Repeat("─", 60))
	fmt.Printf("Domain: %s\n", result.Domain)
	fmt.Printf("Method: %s\n", strings.ToUpper(result.Method))
	fmt.Printf("%s\n", strings.Repeat("─", 60))

	if result.Available {
		fmt.Printf("\n✓ Domain appears to be available\n")
		if result.Message != "" {
			fmt.Printf("  %s\n", result.Message)
		}
		fmt.Println()
		return nil
	}

	if result.Parsed == nil {
		fmt.Printf("\nNo parsed data available\n\n")
		return nil
	}

	parsed := result.Parsed

	// Domain info
	if parsed.DomainName != "" {
		fmt.Printf("\nDomain Name:     %s\n", parsed.DomainName)
	}

	// Registrar
	if parsed.Registrar != "" {
		fmt.Printf("Registrar:       %s\n", parsed.Registrar)
	}

	// Registrant
	if parsed.Registrant != "" {
		fmt.Printf("Registrant:      %s\n", parsed.Registrant)
	}

	// Dates
	fmt.Println()
	if parsed.CreationDate != "" {
		fmt.Printf("Created:         %s\n", formatDate(parsed.CreationDate))
	}
	if parsed.ExpirationDate != "" {
		fmt.Printf("Expires:         %s\n", formatDate(parsed.ExpirationDate))
	}
	if parsed.LastModified != "" {
		fmt.Printf("Last Modified:   %s\n", formatDate(parsed.LastModified))
	}

	// Status
	if len(parsed.Status) > 0 {
		fmt.Printf("\nStatus:\n")
		for _, status := range parsed.Status {
			fmt.Printf("  • %s\n", status)
		}
	}

	// Nameservers
	if len(parsed.Nameservers) > 0 {
		fmt.Printf("\nNameservers:\n")
		for _, ns := range parsed.Nameservers {
			fmt.Printf("  • %s\n", ns)
		}
	}

	// DNSSEC
	if parsed.DNSSEC != "" {
		fmt.Printf("\nDNSSEC:          %s\n", parsed.DNSSEC)
	}

	// WHOIS Server (for WHOIS lookups)
	if parsed.WhoisServer != "" {
		fmt.Printf("WHOIS Server:    %s\n", parsed.WhoisServer)
	}

	fmt.Println()

	// Raw output if requested
	if p.rawOutput && result.Raw != "" {
		fmt.Printf("%s\n", strings.Repeat("─", 60))
		fmt.Printf("Raw Response:\n")
		fmt.Printf("%s\n", strings.Repeat("─", 60))
		fmt.Println(result.Raw)
	}

	return nil
}

// formatDate attempts to format a date string nicely
func formatDate(date string) string {
	// Just return the date as-is for now
	// Could enhance with time.Parse to reformat
	return date
}
