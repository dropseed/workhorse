package main

import (
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use: "workhorse",
	// Version: version.WithMeta,
	// PersistentPreRun: func(cmd *cobra.Command, args []string) {
	// 	if verbose {
	// 		output.Verbosity = 1
	// 	}
	// },
}

func init() {
	// rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
