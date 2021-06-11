package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/subosito/gotenv"

	"github.com/Bnei-Baruch/galaxy-monitor/common"
)

var RootCmd = &cobra.Command{
	Use:   "galaxy-monitor",
	Short: "Galaxy user monitor",
}

func init() {
	cobra.OnInitialize(initConfig)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func initConfig() {
	gotenv.Load()
	common.Init()
}
