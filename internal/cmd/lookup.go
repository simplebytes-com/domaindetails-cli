package cmd

import (
	"fmt"
	"strings"

	"github.com/simplebytes-com/domaindetails-cli/internal/output"
	"github.com/simplebytes-com/domaindetails-cli/internal/rdap"
	"github.com/simplebytes-com/domaindetails-cli/internal/whois"
	"github.com/spf13/cobra"
)

var lookupCmd = &cobra.Command{
	Use:   "lookup <domain>",
	Short: "Look up domain registration info (RDAP preferred, WHOIS fallback)",
	Long: `Performs a domain lookup using RDAP first, falling back to WHOIS if needed.

RDAP (Registration Data Access Protocol) is the modern replacement for WHOIS,
providing structured JSON responses. Most major TLDs now support RDAP.

Examples:
  domaindetails lookup example.com
  domaindetails lookup google.co.uk --json
  domaindetails lookup github.io --raw`,
	Args: cobra.ExactArgs(1),
	RunE: runLookup,
}

func init() {
	rootCmd.AddCommand(lookupCmd)
}

func runLookup(cmd *cobra.Command, args []string) error {
	domain := strings.ToLower(strings.TrimSpace(args[0]))

	if !isValidDomain(domain) {
		return fmt.Errorf("invalid domain format: %s", domain)
	}

	if verbose {
		fmt.Printf("Looking up domain: %s\n", domain)
	}

	// Try RDAP first
	rdapClient := rdap.NewClient(verbose)
	result, err := rdapClient.Lookup(domain)

	if err != nil {
		if verbose {
			fmt.Printf("RDAP lookup failed: %v, falling back to WHOIS\n", err)
		}

		// Fall back to WHOIS
		whoisClient := whois.NewClient(verbose)
		result, err = whoisClient.Lookup(domain)

		if err != nil {
			return fmt.Errorf("lookup failed: %v", err)
		}
	}

	// Output results
	printer := output.NewPrinter(jsonOutput, rawOutput)
	return printer.Print(result)
}

func isValidDomain(domain string) bool {
	// Basic domain validation
	if len(domain) < 3 || len(domain) > 253 {
		return false
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
		// Check for valid characters
		for i, c := range part {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || (c == '-' && i > 0 && i < len(part)-1)) {
				return false
			}
		}
	}

	return true
}
