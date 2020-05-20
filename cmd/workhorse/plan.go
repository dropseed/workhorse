package main

import (
	"time"

	"github.com/dropseed/workhorse/internal/scripts"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Create and save a plan",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		planName := time.Now().UTC().Format(time.RFC3339)
		if _, err := scripts.CreatePlan(args[0], planName); err != nil {
			printErrAndExitFailure(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}
