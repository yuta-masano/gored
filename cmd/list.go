package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list projects in your config file",
	Run: func(cmd *cobra.Command, args []string) {
		for _, v := range cfg.Projects {
			fmt.Println(v)
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
