// Package cmd contains all the commands to run and operate with containerz.
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	addr string
)

// RootCmd is the cmd entrypoint for all containerz commands.
var RootCmd = &cobra.Command{
	Use:   "containerz",
	Short: "Containerz suite of CLI tools",
	Run: func(command *cobra.Command, args []string) {
		command.HelpFunc()(command, args)
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&addr, "addr", ":9999", "Containerz listen port.")
}
