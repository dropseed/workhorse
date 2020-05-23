package github

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/dropseed/workhorse/internal/commands"
)

type PullStep struct {
	AddLabels *AddLabels `yaml:"add_labels,omitempty" json:"add_labels,omitempty" mapstructure:"add_labels"`

	// Generic
	Sleep *commands.Sleep `yaml:"sleep,omitempty" json:"sleep,omitempty" mapstructure:"sleep"`
}

type Pulls struct {
	Search string `yaml:"search" json:"search" mapstructure:"search"`
	// filter
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

func parseIssueTarget(s string) (string, string, int) {
	repoRe := regexp.MustCompile(`/([^/]+)/([^/]+)/(pull|issue)/(\d+)`)
	sub := repoRe.FindAllStringSubmatch(s, -1)
	owner := sub[0][1]
	repo := sub[0][2]
	number, _ := strconv.Atoi(sub[0][3])

	return owner, repo, number
}
