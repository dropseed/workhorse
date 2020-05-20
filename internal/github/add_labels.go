package github

import (
	"context"
	"errors"
)

type AddLabels struct {
}

func (cmd *AddLabels) Run(gh *GitHub, owner string, repo string, number int, args ...interface{}) error {
	labels := []string{}
	for _, arg := range args {
		labels = append(labels, arg.(string))
	}
	_, _, err := gh.client.Issues.AddLabelsToIssue(context.Background(), owner, repo, number, labels)
	return err
}

func (cmd *AddLabels) Validate(args ...interface{}) error {
	if len(args) < 1 {
		return errors.New("Should have at least one label")
	}
	for _, arg := range args {
		if _, ok := arg.(string); !ok {
			return errors.New("Arg is not a string")
		}
	}
	return nil
}
