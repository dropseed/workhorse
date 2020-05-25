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
	Search   string      `yaml:"search" json:"search" mapstructure:"search"`
	Filter   *PullFilter `yaml:"filter" json:"filter" mapstructure:"filter"`
	Markdown string      `yaml:"markdown" json:"markdown" mapstructure:"markdown"`
	Steps    []*PullStep `yaml:"steps" json:"steps" mapstructure:"steps"`
	objs     map[string]*github.PullRequest
}

func (pulls *Pulls) Validate() error {
	if pulls.Markdown != "" {
		if _, err := getTemplate(pulls.Markdown); err != nil {
			return err
		}
	}

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
			pull, err := pulls.getOrFetchPull(url, client)
			if err != nil {
				return nil, err
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

func (pulls *Pulls) targetsAsMarkdown(urls []string, client *github.Client) ([]string, error) {
	md := []string{}
	for _, url := range urls {
		pull, err := pulls.getOrFetchPull(url, client)
		if err != nil {
			return nil, err
		}
		// template.New("pull").Parse()
		markdown, err := toMarkdown(pulls.Markdown, pull)
		if err != nil {
			return nil, err
		}
		md = append(md, markdown)
	}
	return md, nil
}

func (pulls *Pulls) getOrFetchPull(url string, client *github.Client) (*github.PullRequest, error) {
	if pulls.objs == nil {
		pulls.objs = map[string]*github.PullRequest{}
	}

	if cached, ok := pulls.objs[url]; ok {
		return cached, nil
	}

	owner, repo, number := parseIssueTarget(url)
	pull, _, ghErr := client.PullRequests.Get(context.Background(), owner, repo, number)
	if ghErr != nil {
		return nil, ghErr
	}

	pulls.objs[url] = pull

	return pull, nil
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
