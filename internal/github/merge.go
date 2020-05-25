package github

import (
	"context"

	"github.com/google/go-github/v31/github"
)

type Merge struct {
	// Message string `yaml:"message" json:"message" mapstructure:"message"`
	Method string `yaml:"method" json:"method" mapstructure:"method"`
}

func (cmd *Merge) Run(target string) error {
	// pull, err := getOrFetchPull(target)
	owner, repo, number := parseIssueTarget(target)
	opts := &github.PullRequestOptions{
		// CommitTitle: pull.GetTitle(), // TODO could be an option
		// SHA:         pull.Head.GetSHA(),
		MergeMethod: cmd.Method,
	}
	_, _, err := getClient().PullRequests.Merge(context.Background(), owner, repo, number, "", opts)
	return err
}

func (cmd *Merge) Validate() error {
	return nil
}
