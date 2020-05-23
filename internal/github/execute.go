package github

import (
	"fmt"

	"github.com/dropseed/workhorse/internal/commands"
)

func (config *Config) ExecuteTargets(targets []string) error {
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
