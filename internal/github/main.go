package github

import (
	"context"
	"os"

	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
)

var client *github.Client

func getToken() string {
	if s := os.Getenv("WORKHORSE_TOKEN"); s != "" {
		return s
	}
	if s := os.Getenv("GITHUB_TOKEN"); s != "" {
		return s
	}
	return ""
}

func getClient() *github.Client {
	if client != nil {
		return client
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: getToken()},
	)
	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)
	return client
}
