package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Bnei-Baruch/galaxy-monitor/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of galaxy-monitor",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Galaxy Monitor version %s\n", version.Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
