package cmd

import (
	"fmt"
	"strings"

	"github.com/simplebytes-com/domaindetails-cli/internal/output"
	"github.com/simplebytes-com/domaindetails-cli/internal/rdap"
	"github.com/spf13/cobra"
)

var rdapCmd = &cobra.Command{
	Use:   "rdap <domain>",
	Short: "Look up domain using RDAP only",
	Long: `Performs a domain lookup using RDAP (Registration Data Access Protocol) only.

RDAP is the modern replacement for WHOIS, providing structured JSON responses
with better internationalization support and standardized output.

Examples:
  domaindetails rdap example.com
  domaindetails rdap google.co.uk --json
  domaindetails rdap github.io --raw`,
	Args: cobra.ExactArgs(1),
	RunE: runRdap,
}

func init() {
	rootCmd.AddCommand(rdapCmd)
}

func runRdap(cmd *cobra.Command, args []string) error {
	domain := strings.ToLower(strings.TrimSpace(args[0]))

	if !isValidDomain(domain) {
		return fmt.Errorf("invalid domain format: %s", domain)
	}

	if verbose {
		fmt.Printf("RDAP lookup for domain: %s\n", domain)
	}

	rdapClient := rdap.NewClient(verbose)
	result, err := rdapClient.Lookup(domain)

	if err != nil {
		return fmt.Errorf("RDAP lookup failed: %v", err)
	}

	printer := output.NewPrinter(jsonOutput, rawOutput)
	return printer.Print(result)
}
