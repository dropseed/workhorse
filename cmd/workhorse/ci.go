package main

import (
	"fmt"

	"github.com/dropseed/workhorse/internal/git"
	"github.com/dropseed/workhorse/internal/github"
	"github.com/dropseed/workhorse/internal/scripts"
	"github.com/dropseed/workhorse/internal/utils"
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

		base := "master"
		branch := git.CleanBranchName(plan.Script)

		if len(plan.Targets) == 0 {
			println("No targets found")

			owner, repo := github.OwnerRepoFromRemote()
			pr, _ := github.FindPullRequest(owner, repo, base, branch)
			if pr != nil {
				println("Closing open pull request ", pr.GetHTMLURL())
				if err := github.ClosePullRequest(pr); err != nil {
					printErrAndExitFailure(err)
				}
			}

			return
		}

		// should always go off of latest master
		// so delete if exists
		// then create and run plan (will increment plan number too if something else merged - conflicts ARE conflicts and you want things to run in order and force them updating)
		fmt.Printf("Branch: %s\n", branch)
		if err := git.CreateBranch(branch); err != nil {
			if err := git.DeleteBranch(branch); err != nil {
				printErrAndExitFailure(err)
			}
			if err := git.CreateBranch(branch); err != nil {
				printErrAndExitFailure(err)
			}
		}

		planSlug := plan.GetSlug()
		title := fmt.Sprintf("%s: %s", planSlug, utils.ExtensionlessBasename(plan.Script))

		git.Commit(plan.GetPath(), title)
		git.Push(branch)

		// TODO ideally we would only push if there is a difference, not force push every time

		body := fmt.Sprintf("Merging this PR will run %s on the following PRs:\n\n", plan.Script)
		lines, mdErr := plan.TargetsAsMarkdown()
		if mdErr != nil {
			printErrAndExitFailure(mdErr)
		}
		for _, line := range lines {
			body = body + "- " + line + "\n"
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
		// need a SKIP option based on msg or something (how else do you commit a plan you ran manually - or you don't!)
		lastFiles := git.LastCommitFilesAdded(scripts.GetPlansDir())
		if len(lastFiles) > 0 {

			println("Plans to execute:")
			for _, path := range lastFiles {
				println("  " + path)
			}

			for _, path := range lastFiles {
				fmt.Printf("Executing plan: %s", path)
				if err := scripts.ExecutePlan(path); err != nil {
					printErrAndExitFailure(err)
				}
			}
		} else {
			println("No plans in last commit to execute")
		}
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
