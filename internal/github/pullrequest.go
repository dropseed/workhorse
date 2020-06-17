package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/dropseed/workhorse/internal/git"
	"github.com/google/go-github/v31/github"
)

func OwnerRepoFromRemote() (string, string) {
	remote := git.Remote()
	parts := strings.Split(remote, "/")
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]
	if strings.HasSuffix(repo, ".git") {
		repo = repo[:len(repo)-4]
	}
	return owner, repo
}

func PullRequest(base, head, title, body string) (*github.PullRequest, error) {
	ctx := context.Background()
	client := getClient()

	owner, repo := OwnerRepoFromRemote()

	pull := &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Base:  &base,
		Head:  &head,
	}
	pr, _, err := client.PullRequests.Create(ctx, owner, repo, pull)
	if err != nil {
		// if strings.Index(string(resp.Body), "pull request already exists") != -1 {
		// just assume exists err for now
		existing, err := FindPullRequest(owner, repo, base, head)
		if err != nil {
			return nil, err
		}

		existing.Title = &title
		existing.Body = &body
		pr, _, err = client.PullRequests.Edit(ctx, owner, repo, existing.GetNumber(), existing)
		if err != nil {
			return nil, err
		}
	}

	return pr, nil
}

func FindPullRequest(owner, repo, base, head string) (*github.PullRequest, error) {
	ctx := context.Background()
	opt := &github.PullRequestListOptions{
		State: "open",
		Base:  base,
		Head:  owner + ": " + head, // needs to specify user/org
	}
	prs, _, err := getClient().PullRequests.List(ctx, owner, repo, opt)
	if err != nil {
		return nil, err
	}

	if len(prs) != 1 {
		extra := ""
		for _, pr := range prs {
			extra = extra + "\n- " + pr.GetHTMLURL()
		}
		return nil, fmt.Errorf("Found %d matches for existing pull request\n%s", len(prs), extra)
	}

	return prs[0], nil
}

func ClosePullRequest(pr *github.PullRequest) error {
	state := "closed"
	pr.State = &state
	_, _, err := getClient().PullRequests.Edit(context.Background(), pr.GetBase().GetUser().GetLogin(), pr.GetBase().GetRepo().GetName(), pr.GetNumber(), pr)
	return err
}
