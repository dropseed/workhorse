package github

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-github/v31/github"
)

func (config *Config) GetTargets() ([]string, error) {
	if config.Pulls != nil && config.Pulls.Search != "" {
		query := config.Pulls.Search
		if strings.Index(query, "is:pr") == -1 {
			query = "is:pr " + query
		}
		return config.SearchIssues(query)
	}

	return nil, errors.New("Unknown search situation")
}

func (config *Config) SearchIssues(query string) ([]string, error) {
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "created",
		Order:       "desc",
	}

	var allIssues []*github.Issue
	for {
		result, resp, err := config.client.Search.Issues(context.Background(), query, opt)
		if err != nil {
			return nil, err
		}
		allIssues = append(allIssues, result.Issues...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	urls := []string{}

	for _, i := range allIssues {
		url := i.GetHTMLURL()
		urls = append(urls, url)
		fmt.Printf("%s\n", url)
	}

	// sort for diff consistency
	sort.Strings(urls)

	return urls, nil
}
