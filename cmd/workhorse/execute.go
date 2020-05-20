package main

import (
	"github.com/dropseed/workhorse/internal/scripts"
	"github.com/spf13/cobra"
)

var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute a plan by name or path",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := scripts.ExecutePlan(args[0]); err != nil {
			printErrAndExitFailure(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(executeCmd)
}
