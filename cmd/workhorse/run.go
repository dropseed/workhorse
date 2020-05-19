package main

import (
	"github.com/dropseed/workhorse/internal/scripts"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := scripts.RunScript(args[0]); err != nil {
			printErrAndExitFailure(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
