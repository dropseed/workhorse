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

var pullsCache map[string]*github.PullRequest

type PullStep struct {
	AddLabels    *AddLabels    `yaml:"add_labels,omitempty" json:"add_labels,omitempty" mapstructure:"add_labels,omitempty"`
	RemoveLabel  *RemoveLabel  `yaml:"remove_label,omitempty" json:"remove_label,omitempty" mapstructure:"remove_label,omitempty"`
	Close        *Close        `yaml:"close,omitempty" json:"close,omitempty" mapstructure:"close,omitempty"`
	Merge        *Merge        `yaml:"merge,omitempty" json:"merge,omitempty" mapstructure:"merge,omitempty"`
	DeleteBranch *DeleteBranch `yaml:"delete_branch,omitempty" json:"delete_branch,omitempty" mapstructure:"delete_branch,omitempty"`

	// Generic
	Sleep *commands.Sleep `yaml:"sleep,omitempty" json:"sleep,omitempty" mapstructure:"sleep,omitempty"`
}

type PullFilter struct {
	Mergeable      *bool   `yaml:"mergeable,omitempty" json:"mergeable,omitempty" mapstructure:"mergeable,omitempty"`
	MergeableState *string `yaml:"mergeable_state,omitempty" json:"mergeable_state,omitempty" mapstructure:"mergeable_state,omitempty"`
}

type Pulls struct {
	Search   string      `yaml:"search,omitempty" json:"search,omitempty" mapstructure:"search,omitempty"`
	Filter   *PullFilter `yaml:"filter,omitempty" json:"filter,omitempty" mapstructure:"filter,omitempty"`
	Markdown string      `yaml:"markdown,omitempty" json:"markdown,omitempty" mapstructure:"markdown,omitempty"`
	Steps    []*PullStep `yaml:"steps,omitempty" json:"steps,omitempty" mapstructure:"steps,omitempty"`
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

func (pulls *Pulls) getTargets() ([]string, error) {
	query := pulls.Search
	if strings.Index(query, "is:pr") == -1 {
		query = "is:pr " + query
	}

	pullUrls, err := searchIssues(query)
	if err != nil {
		return nil, err
	}

	filteredUrls := []string{}

	if pulls.Filter != nil {
		for _, url := range pullUrls {
			pull, err := getOrFetchPull(url)
			if err != nil {
				return nil, err
			}

			match := true

			if pulls.Filter.Mergeable != nil && *pulls.Filter.Mergeable != pull.GetMergeable() {
				match = false
			}

			if pulls.Filter.MergeableState != nil && *pulls.Filter.MergeableState != pull.GetMergeableState() {
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

func (pulls *Pulls) targetsAsMarkdown(urls []string) ([]string, error) {
	md := []string{}
	for _, url := range urls {
		pull, err := getOrFetchPull(url)
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

func getOrFetchPull(url string) (*github.PullRequest, error) {
	if pullsCache == nil {
		pullsCache = map[string]*github.PullRequest{}
	}

	if cached, ok := pullsCache[url]; ok {
		return cached, nil
	}

	owner, repo, number := parseIssueTarget(url)
	pull, _, ghErr := getClient().PullRequests.Get(context.Background(), owner, repo, number)
	if ghErr != nil {
		return nil, ghErr
	}

	pullsCache[url] = pull

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
