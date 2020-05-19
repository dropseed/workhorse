package scripts

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"

	"github.com/dropseed/workhorse/internal/config"
	"github.com/dropseed/workhorse/internal/github"
	"github.com/mitchellh/mapstructure"
)

type Plan struct {
	// version of release that ran it?
	Script  string         `json:"script"`
	Targets []string       `json:"targets"`
	Config  *config.Config `json:"config"`
	client  *github.GitHub
}

func NewPlan(script string, config *config.Config) (*Plan, error) {
	client, err := github.NewGitHub()
	if err != nil {
		return nil, err
	}

	return &Plan{
		Script:  script,
		Targets: []string{},
		Config:  config,
		client:  client,
	}, nil
}

func NewPlanFromPath(path string) (*Plan, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	return NewPlanFromReader(f)
}

func NewPlanFromReader(reader io.Reader) (*Plan, error) {
	temp := map[string]interface{}{}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&temp); err != nil {
		return nil, err
	}

	return newPlanFromMap(temp)
}

func newPlanFromMap(m map[string]interface{}) (*Plan, error) {
	plan := &Plan{}

	mapDecoderConfig := mapstructure.DecoderConfig{
		Result:      plan,
		ErrorUnused: true,
	}
	mapDecoder, err := mapstructure.NewDecoder(&mapDecoderConfig)
	if err != nil {
		return nil, err
	}

	if err = mapDecoder.Decode(m); err != nil {
		return nil, err
	}

	client, err := github.NewGitHub()
	if err != nil {
		return nil, err
	}
	plan.client = client

	return plan, nil
}

func (p *Plan) Validate() error {
	for _, s := range p.Config.Steps {
		if err := p.client.ValidateCommand(s.Run, s.Args...); err != nil {
			fmt.Printf("%+v\n%+v\n", s.Run, s.Args)
			return err
		}
	}
	return nil
}

func (p *Plan) Load() error {
	fmt.Printf("Query:\n%s\n\n", p.Config.Issues)

	issues, err := p.client.Search(p.Config.Issues)
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

func (p *Plan) GetPath(name string) string {
	return path.Join("workhorse", "plans", fmt.Sprintf("%s.json", name))
}

func (p *Plan) Save(name string) (string, error) {
	out, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	out = append(out, "\n"...)
	path := p.GetPath(name)
	if err := ioutil.WriteFile(path, out, 0644); err != nil {
		panic(err)
	}
	return path, nil
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
