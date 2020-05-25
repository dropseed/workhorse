package github

import (
	"context"

	"github.com/google/go-github/v31/github"
)

type Close struct {
}

func (cmd *Close) Run(target string) error {
	owner, repo, number := parseIssueTarget(target)
	client := newClient()
	state := "closed"
	issue := &github.IssueRequest{
		State: &state,
	}
	_, _, err := client.Issues.Edit(context.Background(), owner, repo, number, issue)
	return err
}

func (cmd *Close) Validate() error {
	return nil
}
