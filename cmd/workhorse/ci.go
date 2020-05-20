package main

import (
	"fmt"

	"github.com/dropseed/workhorse/internal/git"
	"github.com/dropseed/workhorse/internal/github"
	"github.com/dropseed/workhorse/internal/meta"
	"github.com/dropseed/workhorse/internal/scripts"
	"github.com/spf13/cobra"
)

var force bool

var ciPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !force && git.IsDirty() {
			printErrAndExitFailure(fmt.Errorf("Git status must be clean first\n\n%s", git.Status()))
		}

		plan, err := scripts.CreatePlan(args[0])
		if err != nil {
			printErrAndExitFailure(err)
		}

		if len(plan.Targets) == 0 {
			println("No targets found for plan")
			return
		}

		branch := git.CleanBranchName(plan.Script)
		fmt.Printf("Branch: %s\n", branch)

		base := "master"

		// should always go off of latest master
		// so delete if exists
		// then create and run plan (will increment plan number too if something else merged - conflicts ARE conflicts and you want things to run in order and force them updating)
		if err := git.CreateBranch(branch); err != nil {
			if err := git.DeleteBranch(branch); err != nil {
				printErrAndExitFailure(err)
			}
			if err := git.CreateBranch(branch); err != nil {
				printErrAndExitFailure(err)
			}
		}

		planSlug := plan.GetSlug()
		git.Commit(plan.GetPath(), fmt.Sprintf("Create %s plan %s", meta.AppName, planSlug))
		git.Push(branch)

		title := fmt.Sprintf("%s: %s", planSlug, plan.Script)
		body := fmt.Sprintf("Merging this PR will run %s on the following PRs:\n\n", plan.Script)
		for _, target := range plan.Targets {
			body = body + "- " + target + "\n"
		}

		if pr, err := github.PullRequest(base, branch, title, body); err != nil {
			printErrAndExitFailure(err)
		} else {
			println(pr.GetHTMLURL())
		}

		git.Checkout("-")
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
