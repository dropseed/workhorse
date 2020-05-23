package github

import (
	"io"
	"os"

	"github.com/google/go-github/v31/github"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Pulls *Pulls `yaml:"pulls" json:"pulls" mapstructure:"pulls"`
	// Issues
	// Repos
	client *github.Client
}

func (config *Config) Validate() error {

	// should only have pulls, issues, or repos
	if config.Pulls != nil {
		if err := config.Pulls.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func NewConfigFromPath(path string) (*Config, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	return NewConfigFromReader(f)
}

func NewConfigFromReader(reader io.Reader) (*Config, error) {
	temp := map[string]interface{}{}
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&temp); err != nil {
		return nil, err
	}

	return newConfigFromMap(temp)
}

func newConfigFromMap(m map[string]interface{}) (*Config, error) {
	config := &Config{}

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

	config.client = newClient()

	return config, nil
}
