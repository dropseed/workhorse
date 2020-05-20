package github

import (
	"context"
	"errors"
)

type RemoveLabel struct {
}

func (cmd *RemoveLabel) Run(gh *GitHub, owner string, repo string, number int, args ...interface{}) error {
	label := args[0].(string)
	_, err := gh.client.Issues.RemoveLabelForIssue(context.Background(), owner, repo, number, label)
	return err
}

func (cmd *RemoveLabel) Validate(args ...interface{}) error {
	if len(args) != 1 {
		return errors.New("Should have at exactly one label")
	}
	for _, arg := range args {
		if _, ok := arg.(string); !ok {
			return errors.New("Arg is not a string")
		}
	}
	return nil
}
