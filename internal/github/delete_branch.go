package github

import (
	"context"
	"io/ioutil"
	"strings"
)

type DeleteBranch struct {
	Name       string `yaml:"name,omitempty" json:"name,omitempty" mapstructure:"name,omitempty"`
	AllowError string `yaml:"allow_error,omitempty" json:"allow_error,omitempty" mapstructure:"allow_error,omitempty"`
}

func (cmd *DeleteBranch) Run(target string) error {
	pull, err := getOrFetchPull(target)
	owner, repo, _ := parseIssueTarget(target)
	ref := "heads/" + pull.GetHead().GetRef() // use PR branch by default
	if cmd.Name != "" {
		ref = "heads/" + cmd.Name
	}

	resp, err := getClient().Git.DeleteRef(context.Background(), owner, repo, ref)

	if err != nil && cmd.AllowError != "" {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if strings.Index(string(body), cmd.AllowError) != -1 {
			return nil
		}
	}

	return err
}

func (cmd *DeleteBranch) Validate() error {
	return nil
}
