package github

import (
	"errors"
	"fmt"
	"regexp"
	"sort"

	"github.com/dropseed/workhorse/internal/commands"
	"github.com/google/go-github/v31/github"
)

type RepoStep struct {
	Edit *RepoEdit `yaml:"edit,omitempty" json:"edit,omitempty" mapstructure:"edit,omitempty"`

	// Generic
	Sleep *commands.Sleep `yaml:"sleep,omitempty" json:"sleep,omitempty" mapstructure:"sleep,omitempty"`
}

type RepoFilter struct {
	DeleteBranchOnMerge *bool `yaml:"delete_branch_on_merge,omitempty" json:"delete_branch_on_merge,omitempty" mapstructure:"delete_branch_on_merge,omitempty"`
}

type Repos struct {
	Search string      `yaml:"search,omitempty" json:"search,omitempty" mapstructure:"search,omitempty"`
	Filter *RepoFilter `yaml:"filter,omitempty" json:"filter,omitempty" mapstructure:"filter,omitempty"`
	// markdown
	Steps []*RepoStep `yaml:"steps,omitempty" json:"steps,omitempty" mapstructure:"steps,omitempty"`
}

func (repos *Repos) Validate() error {
	for _, step := range repos.Steps {
		cmds := commands.CommandStructFields(step)
		if len(cmds) != 1 {
			return errors.New("Each step should have 1 command")
		}
		for _, command := range cmds {
			if err := command.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (repos *Repos) getTargets(client *github.Client) ([]string, error) {
	query := repos.Search

	allRepos, err := searchRepos(client, query)
	if err != nil {
		return nil, err
	}

	filteredRepos := []*github.Repository{}

	if repos.Filter != nil {
		for _, repo := range allRepos {
			match := true

			if repos.Filter.DeleteBranchOnMerge != nil && *repos.Filter.DeleteBranchOnMerge != repo.GetDeleteBranchOnMerge() {
				match = false
			}

			if match {
				filteredRepos = append(filteredRepos, repo)
			}
		}
	} else {
		filteredRepos = allRepos
	}

	urls := []string{}

	for _, i := range filteredRepos {
		url := i.GetHTMLURL()
		urls = append(urls, url)
		fmt.Printf("%s\n", url)
	}
	sort.Strings(urls) // sort for diff consistency

	return urls, nil
}

func parseRepoTarget(s string) (string, string) {
	repoRe := regexp.MustCompile(`/([^/]+)/([^/]+)$`)
	sub := repoRe.FindAllStringSubmatch(s, -1)
	owner := sub[0][1]
	repo := sub[0][2]

	return owner, repo
}
