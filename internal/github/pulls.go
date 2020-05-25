package github

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropseed/workhorse/internal/commands"
	"github.com/google/go-github/v31/github"
)

type PullStep struct {
	AddLabels   *AddLabels   `yaml:"add_labels,omitempty" json:"add_labels,omitempty" mapstructure:"add_labels,omitempty"`
	RemoveLabel *RemoveLabel `yaml:"remove_label,omitempty" json:"remove_label,omitempty" mapstructure:"remove_label,omitempty"`
	Close       *Close       `yaml:"close,omitempty" json:"close,omitempty" mapstructure:"close,omitempty"`

	// Generic
	Sleep *commands.Sleep `yaml:"sleep,omitempty" json:"sleep,omitempty" mapstructure:"sleep,omitempty"`
}

type PullFilter struct {
	Mergeable *bool `yaml:"mergeable,omitempty" json:"mergeable,omitempty" mapstructure:"mergeable,omitempty"`
}

type Pulls struct {
	Search string      `yaml:"search" json:"search" mapstructure:"search"`
	Filter *PullFilter `yaml:"filter" json:"filter" mapstructure:"filter"`
	// markdown
	Steps []*PullStep `yaml:"steps" json:"steps" mapstructure:"steps"`
}

func (pulls *Pulls) Validate() error {
	for _, step := range pulls.Steps {
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

func (pulls *Pulls) getTargets(client *github.Client) ([]string, error) {
	query := pulls.Search
	if strings.Index(query, "is:pr") == -1 {
		query = "is:pr " + query
	}

	pullUrls, err := searchIssues(client, query)
	if err != nil {
		return nil, err
	}

	filteredUrls := []string{}

	if pulls.Filter != nil {
		for _, url := range pullUrls {
			owner, repo, number := parseIssueTarget(url)
			pull, _, ghErr := client.PullRequests.Get(context.Background(), owner, repo, number)
			if ghErr != nil {
				return nil, ghErr
			}

			match := true

			if pulls.Filter.Mergeable != nil && *pulls.Filter.Mergeable != pull.GetMergeable() {
				match = false
			}

			if match {
				filteredUrls = append(filteredUrls, url)
			}
		}
	} else {
		filteredUrls = pullUrls
	}

	return filteredUrls, nil
}

func parseIssueTarget(s string) (string, string, int) {
	repoRe := regexp.MustCompile(`/([^/]+)/([^/]+)/(pull|issue)/(\d+)$`)
	sub := repoRe.FindAllStringSubmatch(s, -1)
	owner := sub[0][1]
	repo := sub[0][2]
	number, err := strconv.Atoi(sub[0][4])
	if err != nil {
		panic(err)
	}

	return owner, repo, number
}
