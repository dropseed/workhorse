package github

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
)

type GitHub struct {
	client   *github.Client
	commands map[string]Command
}

func getToken() string {
	if s := os.Getenv("WORKHORSE_TOKEN"); s != "" {
		return s
	}
	if s := os.Getenv("GITHUB_TOKEN"); s != "" {
		return s
	}
	return ""
}

func newClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: getToken()},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func NewGitHub() (*GitHub, error) {
	client := newClient()
	gh := &GitHub{
		client:   client,
		commands: map[string]Command{},
	}

	gh.RegisterCommand("remove_label", &RemoveLabel{})
	gh.RegisterCommand("add_labels", &AddLabels{})
	gh.RegisterCommand("sleep", &Sleep{})

	return gh, nil
}

func (gh *GitHub) RegisterCommand(name string, command Command) error {
	gh.commands[name] = command
	return nil
}

func (gh *GitHub) RunCommand(name string, owner string, repo string, number int, args ...interface{}) error {
	cmd, ok := gh.commands[name]
	if !ok {
		return fmt.Errorf("%s command doesn't exist", name)
	}
	return cmd.Run(gh, owner, repo, number, args...)
}

func (gh *GitHub) ValidateCommand(name string, args ...interface{}) error {
	cmd, ok := gh.commands[name]
	if !ok {
		return fmt.Errorf("%s command doesn't exist", name)
	}
	return cmd.Validate(args...)
}

func (gh *GitHub) Search(query string) ([]*github.Issue, error) {
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "created",
		Order:       "desc",
	}

	var allIssues []*github.Issue
	for {
		result, resp, err := gh.client.Search.Issues(context.Background(), query, opt)
		if err != nil {
			return nil, err
		}
		allIssues = append(allIssues, result.Issues...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allIssues, nil
}
