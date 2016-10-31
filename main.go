package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version bool

	// These values are embedded when building.
	buildVersion string
	buildDate    string
	buildWith    string
)

var rootCmd = &cobra.Command{
	Use:   "gored project_id",
	Short: `gored adds a new issue using your clipboard text, returns the added issue pages's title and URL.`,
	RunE:  runGored,
}

func init() {
	rootCmd.Flags().BoolVarP(&version, "version", "v", false,
		"show program's version number and exit")
}

func runGored(cmd *cobra.Command, args []string) error {
	if version {
		fmt.Printf("version: %s\nbuild at: %s\nwith: %s\n",
			buildVersion, buildDate, buildWith)
		return nil
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
