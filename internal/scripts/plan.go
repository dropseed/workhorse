package scripts

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/dropseed/workhorse/internal/config"
	"github.com/dropseed/workhorse/internal/github"
)

type Plan struct {
	// version of release that ran it?
	name    string
	Script  string         `json:"script"`
	Targets []string       `json:"targets"`
	Config  *config.Config `json:"config"`
	client  *github.GitHub
}

func NewPlan(name string, script string, config *config.Config) (*Plan, error) {
	client, err := github.NewGitHub()
	if err != nil {
		return nil, err
	}

	return &Plan{
		name:    name,
		Script:  script,
		Targets: []string{},
		Config:  config,
		client:  client,
	}, nil
}

func (p *Plan) Validate() error {
	for _, s := range p.Config.Steps {
		if err := p.client.ValidateCommand(s.Run, s.Args...); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plan) Load() error {
	issues, err := p.client.Search(p.Config.Search.Issues.Q)
	if err != nil {
		return err
	}
	for _, i := range issues {
		url := i.GetHTMLURL()
		p.Targets = append(p.Targets, url)
		fmt.Printf("%s\n", url)
	}
	return nil
}

func (p *Plan) Save() {

}

func (p *Plan) Execute() error {
	repoRe := regexp.MustCompile(`/([^/]+)/([^/]+)/pull/(\d+)`)

	for _, target := range p.Targets {
		fmt.Printf("%s\n", target)
		for _, s := range p.Config.Steps {
			fmt.Printf("  %s\n", s.Run)

			sub := repoRe.FindAllStringSubmatch(target, -1)
			owner := sub[0][1]
			repo := sub[0][2]
			number, _ := strconv.Atoi(sub[0][3])

			if err := p.client.RunCommand(s.Run, owner, repo, number, s.Args...); err != nil {
				return err
			}
		}
		break
	}

	return nil
}
