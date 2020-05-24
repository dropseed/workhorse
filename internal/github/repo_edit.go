package github

import (
	"context"

	"github.com/google/go-github/v31/github"
)

type RepoEdit struct {
	DeleteBranchOnMerge *bool `yaml:"delete_branch_on_merge,omitempty" json:"delete_branch_on_merge,omitempty" mapstructure:"delete_branch_on_merge,omitempty"`
}

func (cmd *RepoEdit) Run(target string) error {
	owner, repo := parseRepoTarget(target)
	client := newClient()
	repoObj := &github.Repository{
		DeleteBranchOnMerge: cmd.DeleteBranchOnMerge,
	}
	_, _, err := client.Repositories.Edit(context.Background(), owner, repo, repoObj)
	return err
}

func (cmd *RepoEdit) Validate() error {
	return nil
}
