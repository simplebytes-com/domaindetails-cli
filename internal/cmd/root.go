// Package cmd provides the CLI commands for domaindetails
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version info set by main
	versionStr string
	commitStr  string
	dateStr    string

	// Global flags
	jsonOutput bool
	rawOutput  bool
	verbose    bool
)

// SetVersionInfo sets version information from build
func SetVersionInfo(version, commit, date string) {
	versionStr = version
	commitStr = commit
	dateStr = date
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(fmt.Sprintf("domaindetails %s (commit: %s, built: %s)\n", version, commit, date))
}

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "domaindetails",
	Short: "Domain RDAP and WHOIS lookup tool",
	Long: `domaindetails - A fast CLI tool for domain registration lookups

Performs RDAP lookups (preferred) with WHOIS fallback for comprehensive
domain registration information. Caches IANA bootstrap data locally for
improved performance.

Built by DomainDetails.com - Privacy-first domain intelligence.

Examples:
  domaindetails lookup example.com
  domaindetails rdap google.com --json
  domaindetails whois github.io --raw

Documentation: https://domaindetails.com/kb/cli
Source: https://github.com/simplebytes-com/domaindetails-cli`,
	Version: versionStr,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVarP(&rawOutput, "raw", "r", false, "Include raw response data")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}
