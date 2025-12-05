package cmd

import (
	"fmt"

	"github.com/simplebytes-com/domaindetails-cli/internal/cache"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the local RDAP bootstrap cache",
	Long: `Manage the local cache of IANA RDAP bootstrap data.

The CLI caches the RDAP bootstrap file from data.iana.org to avoid
repeated network requests. The cache is stored in ~/.domaindetails/

Examples:
  domaindetails cache update    # Force update the cache
  domaindetails cache info      # Show cache status
  domaindetails cache clear     # Clear the cache`,
}

var cacheUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Force update the RDAP bootstrap cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cache.NewCache()
		if err := c.Update(); err != nil {
			return fmt.Errorf("failed to update cache: %v", err)
		}
		fmt.Println("Cache updated successfully")
		return nil
	},
}

var cacheInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show cache status and statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cache.NewCache()
		info, err := c.Info()
		if err != nil {
			return fmt.Errorf("failed to get cache info: %v", err)
		}

		fmt.Printf("Cache directory: %s\n", info.Path)
		fmt.Printf("Last updated:    %s\n", info.LastUpdated.Format("2006-01-02 15:04:05"))
		fmt.Printf("TLDs cached:     %d\n", info.TLDCount)
		fmt.Printf("Cache age:       %s\n", info.Age.Round(1).String())
		fmt.Printf("Cache valid:     %v\n", info.IsValid)

		return nil
	},
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the local cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cache.NewCache()
		if err := c.Clear(); err != nil {
			return fmt.Errorf("failed to clear cache: %v", err)
		}
		fmt.Println("Cache cleared successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheUpdateCmd)
	cacheCmd.AddCommand(cacheInfoCmd)
	cacheCmd.AddCommand(cacheClearCmd)
}
