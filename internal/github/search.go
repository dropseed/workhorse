package github

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/go-github/v31/github"
)

func searchIssues(client *github.Client, query string) ([]string, error) {
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "created",
		Order:       "desc",
	}

	var allIssues []*github.Issue
	for {
		result, resp, err := client.Search.Issues(context.Background(), query, opt)
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

func searchRepos(client *github.Client, query string) ([]*github.Repository, error) {
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "created",
		Order:       "desc",
	}

	var allRepos []*github.Repository
	for {
		result, resp, err := client.Search.Repositories(context.Background(), query, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, result.Repositories...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}
