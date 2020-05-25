package github

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/dropseed/workhorse/internal/commands"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

type GitHubConfig struct {
	Pulls *Pulls `yaml:"pulls,omitempty" json:"pulls,omitempty" mapstructure:"pulls,omitempty"`
	Repos *Repos `yaml:"repos,omitempty" json:"repos,omitempty" mapstructure:"repos,omitempty"`
	// Issues
	// Repos
}

func (config *GitHubConfig) Validate() error {

	// TODO should only have pulls, issues, or repos

	if config.Pulls != nil {
		if err := config.Pulls.Validate(); err != nil {
			return err
		}
	}

	if config.Repos != nil {
		if err := config.Repos.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (config *GitHubConfig) GetTargets() ([]string, error) {
	if config.Pulls != nil && config.Pulls.Search != "" {
		return config.Pulls.getTargets()
	}

	if config.Repos != nil && config.Repos.Search != "" {
		return config.Repos.getTargets()
	}

	return nil, errors.New("Unknown search situation")
}

func (config *GitHubConfig) TargetsAsMarkdown(targets []string) ([]string, error) {
	if config.Pulls != nil && config.Pulls.Markdown != "" {
		return config.Pulls.targetsAsMarkdown(targets)
	}
	// TODO repos
	return targets, nil
}

func (config *GitHubConfig) ExecuteTargets(targets []string) error {
	for _, target := range targets {
		fmt.Printf("%s\n", target)

		if config.Pulls != nil {
			for _, s := range config.Pulls.Steps {
				for _, cmd := range commands.CommandStructFields(s) {
					if err := cmd.Run(target); err != nil {
						return err
					}
				}
			}
		}

	}

	return nil
}

func NewConfigFromPath(path string) (*GitHubConfig, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	return newConfigFromReader(f)
}

func newConfigFromReader(reader io.Reader) (*GitHubConfig, error) {
	temp := map[string]interface{}{}
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&temp); err != nil {
		return nil, err
	}

	return newConfigFromMap(temp)
}

func newConfigFromMap(m map[string]interface{}) (*GitHubConfig, error) {
	config := &GitHubConfig{}

	mapDecoderConfig := mapstructure.DecoderConfig{
		Result:      config,
		ErrorUnused: true,
	}
	mapDecoder, err := mapstructure.NewDecoder(&mapDecoderConfig)
	if err != nil {
		return nil, err
	}

	if err = mapDecoder.Decode(m); err != nil {
		return nil, err
	}

	return config, nil
}
