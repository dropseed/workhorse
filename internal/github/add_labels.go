package github

import (
	"context"
	"errors"
)

type AddLabels struct {
	Labels []string `yaml:"labels" json:"labels" mapstructure:"labels"`
}

func (cmd *AddLabels) Run(target string) error {
	owner, repo, number := parseIssueTarget(target)
	_, _, err := getClient().Issues.AddLabelsToIssue(context.Background(), owner, repo, number, cmd.Labels)
	return err
}

func (cmd *AddLabels) Validate() error {
	if len(cmd.Labels) < 1 {
		return errors.New("Should have at least one label")
	}
	return nil
}
