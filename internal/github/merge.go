package github

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/go-github/v31/github"
)

type Merge struct {
	// Message string `yaml:"message" json:"message" mapstructure:"message"`
	Method string `yaml:"method" json:"method" mapstructure:"method"`
	Retry  bool   `yaml:"retry,omitempty" json:"retry,omitempty" mapstructure:"retry,omitempty"`
}

func (cmd *Merge) Run(target string) error {
	// pull, err := getOrFetchPull(target)
	owner, repo, number := parseIssueTarget(target)
	opts := &github.PullRequestOptions{
		// CommitTitle: pull.GetTitle(), // TODO could be an option
		// SHA:         pull.Head.GetSHA(),
		MergeMethod: cmd.Method,
	}
	operation := func() error {
		_, _, err := getClient().PullRequests.Merge(context.Background(), owner, repo, number, "", opts)
		return err
	}

	if cmd.Retry {
		boff := backoff.NewExponentialBackOff()
		boff.MaxElapsedTime = 5 * time.Minute
		err := backoff.Retry(operation, boff)
		return err
	}

	err := operation()
	return err
}

func (cmd *Merge) Validate() error {
	return nil
}
