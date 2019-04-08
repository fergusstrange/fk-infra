package cmd

import (
	"fmt"
	"github.com/infinityworks/fk-infra/util"
	"github.com/spf13/cobra"
	"os"
)

var RootCmd = &cobra.Command{
	Use:   "fk-infra",
	Short: "Create a kubernetes cluster and additional infrastructure to complement",
	Run: func(cmd *cobra.Command, args []string) {
		util.CheckError(cmd.Help())
	},
}

func init() {
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
