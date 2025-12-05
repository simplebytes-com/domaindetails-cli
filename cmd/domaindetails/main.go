// domaindetails - A CLI tool for domain RDAP and WHOIS lookups
//
// Copyright (c) 2024 Simple Bytes LLC
// https://domaindetails.com
//
// This tool provides fast domain registration lookups using RDAP (preferred)
// with WHOIS fallback. It caches the IANA RDAP bootstrap data locally for
// improved performance.

package main

import (
	"fmt"
	"os"

	"github.com/simplebytes-com/domaindetails-cli/internal/cmd"
)

// Version information (set at build time)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
