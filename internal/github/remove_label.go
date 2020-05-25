package github

import (
	"context"
)

type RemoveLabel struct {
	Label string `yaml:"label" json:"label" mapstructure:"label"`
}

func (cmd *RemoveLabel) Run(target string) error {
	owner, repo, number := parseIssueTarget(target)
	_, err := getClient().Issues.RemoveLabelForIssue(context.Background(), owner, repo, number, cmd.Label)
	return err
}

func (cmd *RemoveLabel) Validate() error {
	return nil
}
