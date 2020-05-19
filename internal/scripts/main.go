package scripts

import (
	"fmt"
	"path"

	"github.com/dropseed/workhorse/internal/config"
)

func RunScript(name string) error {
	config, err := config.NewConfigFromPath(path.Join("workhorse", "scripts", fmt.Sprintf("%s.yml", name)))
	if err != nil {
		return err
	}

	plan, err := NewPlan("test", name, config)
	if err != nil {
		return err
	}

	if err := plan.Validate(); err != nil {
		return err
	}

	if err := plan.Load(); err != nil {
		return err
	}

	fmt.Printf("%d targets found\n", len(plan.Targets))
	if len(plan.Targets) < 1 {
		return nil
	}

	// TODO confirm
	plan.Execute()

	// separate plan and apply commands
	// plan saves to json by name
	// apply runs plan

	return nil
}
