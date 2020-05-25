package github

import (
	"context"
)

type DeleteBranch struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty" mapstructure:"name,omitempty"`
}

func (cmd *DeleteBranch) Run(target string) error {
	pull, err := getOrFetchPull(target)
	owner, repo, _ := parseIssueTarget(target)
	ref := pull.GetBase().GetRef()
	if cmd.Name != "" {
		ref = "heads/" + cmd.Name
	}
	_, err = getClient().Git.DeleteRef(context.Background(), owner, repo, ref)
	return err
}

func (cmd *DeleteBranch) Validate() error {
	return nil
}
