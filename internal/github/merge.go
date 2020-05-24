package github

import (
	"context"

	"github.com/google/go-github/v31/github"
)

type Merge struct {
	Message string `yaml:"message" json:"message" mapstructure:"message"`
	Method  string `yaml:"method" json:"method" mapstructure:"method"`
}

func (cmd *Merge) Run(target string) error {
	owner, repo, number := parseIssueTarget(target)
	client := newClient()
	opts := &github.PullRequestOptions{
		CommitTitle: "",
		SHA:         "",
		MergeMethod: cmd.Method,
	}
	_, _, err := client.PullRequests.Merge(context.Background(), owner, repo, number, cmd.Message, opts)
	return err
}

func (cmd *Merge) Validate() error {
	return nil
}
