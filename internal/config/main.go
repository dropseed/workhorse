package config

import (
	"io"
	"os"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

type Step struct {
	Run  string        `yaml:"run" json:"run"`
	Args []interface{} `yaml:"args" json:"args"`
}

type Config struct {
	Issues string  `yaml:"issues" json:"issues"`
	Steps  []*Step `yaml:"steps" json:"steps"`
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

	return config, nil
}
