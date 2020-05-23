package commands

import (
	"errors"
	"time"
)

type Sleep struct {
	Duration float32 `yaml:"duration" json:"duration" mapstructure:"duration"`
}

func (cmd *Sleep) Run(target string) error {
	duration := time.Duration(cmd.Duration)
	time.Sleep(duration * time.Second)
	return nil
}

func (cmd *Sleep) Validate() error {
	if cmd.Duration <= 0 {
		return errors.New("Duration should be positive")
	}
	return nil
}
