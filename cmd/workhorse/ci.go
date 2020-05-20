package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/dropseed/workhorse/internal/git"
	"github.com/dropseed/workhorse/internal/scripts"
	"github.com/spf13/cobra"
)

var force bool

var ciPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// get branch name and checkout or create
		// if exists, then rebase?
		if !force && git.IsDirty() {
			printErrAndExitFailure(errors.New("Git status must be clean first"))
		}

		planName := time.Now().UTC().Format(time.RFC3339)

		plan, err := scripts.CreatePlan(args[0], planName)
		if err != nil {
			printErrAndExitFailure(err)
		}

		branch := git.CleanBranchName(plan.Script)
		fmt.Printf("Branch: %s\n", branch)

		git.Branch(branch)
		// commit
		// open or update PR
	},
}

var ciExecuteCmd = &cobra.Command{
	Use:   "execute",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		// execute plans
		// based purely on git commit? whatever was committed
		// need a SKIP option based on msg or something (how else do you commit a plan you ran manually - or you don't!)
	},
}

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "CI commands",
}

func init() {
	ciPlanCmd.Flags().BoolVarP(&force, "force", "", false, "Force")
	ciCmd.AddCommand(ciPlanCmd)
	ciCmd.AddCommand(ciExecuteCmd)
	rootCmd.AddCommand(ciCmd)
}
