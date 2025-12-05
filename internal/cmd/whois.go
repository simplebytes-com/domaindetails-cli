package cmd

import (
	"fmt"
	"strings"

	"github.com/simplebytes-com/domaindetails-cli/internal/output"
	"github.com/simplebytes-com/domaindetails-cli/internal/whois"
	"github.com/spf13/cobra"
)

var whoisCmd = &cobra.Command{
	Use:   "whois <domain>",
	Short: "Look up domain using WHOIS only",
	Long: `Performs a domain lookup using WHOIS only via the DomainDetails.com API.

WHOIS is the traditional domain registration lookup protocol. This command
queries the DomainDetails.com API which handles the raw WHOIS queries and
parses the responses using the open-source @domaindetails/whois-parser.

Examples:
  domaindetails whois example.com
  domaindetails whois google.co.uk --json
  domaindetails whois github.io --raw`,
	Args: cobra.ExactArgs(1),
	RunE: runWhois,
}

func init() {
	rootCmd.AddCommand(whoisCmd)
}

func runWhois(cmd *cobra.Command, args []string) error {
	domain := strings.ToLower(strings.TrimSpace(args[0]))

	if !isValidDomain(domain) {
		return fmt.Errorf("invalid domain format: %s", domain)
	}

	if verbose {
		fmt.Printf("WHOIS lookup for domain: %s\n", domain)
	}

	whoisClient := whois.NewClient(verbose)
	result, err := whoisClient.Lookup(domain)

	if err != nil {
		return fmt.Errorf("WHOIS lookup failed: %v", err)
	}

	printer := output.NewPrinter(jsonOutput, rawOutput)
	return printer.Print(result)
}
