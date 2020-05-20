package scripts

import (
	"fmt"
	"os"
	"path"

	"github.com/dropseed/workhorse/internal/config"
	"github.com/dropseed/workhorse/internal/meta"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func find(dir, name, extension string) string {
	if fileExists(name) {
		return name
	}
	return path.Join(meta.AppName, dir, fmt.Sprintf("%s.%s", name, extension))
}

func FindScript(name string) string {
	return find("scripts", name, "yml")
}

func FindPlan(name string) string {
	return find("plans", name, "json")
}

// func RunScript(scriptName string) error {

// 	config, err := config.NewConfigFromPath(Find("scripts", scriptName))
// 	if err != nil {
// 		return err
// 	}

// 	plan, err := NewPlan("test", scriptName, config)
// 	if err != nil {
// 		return err
// 	}

// 	if err := plan.Validate(); err != nil {
// 		return err
// 	}

// 	if err := plan.Load(); err != nil {
// 		return err
// 	}

// 	fmt.Printf("%d targets found\n", len(plan.Targets))
// 	if len(plan.Targets) < 1 {
// 		return nil
// 	}

// 	// TODO confirm
// 	plan.Execute()

// 	// separate plan and apply commands
// 	// plan saves to json by scriptName
// 	// apply runs plan

// 	return nil
// }

func CreatePlan(script string) (*Plan, error) {
	scriptPath := FindScript(script)
	config, err := config.NewConfigFromPath(scriptPath)
	if err != nil {
		return nil, err
	}

	plan, err := NewPlan(scriptPath, config)
	if err != nil {
		return nil, err
	}

	if err := plan.Validate(); err != nil {
		return nil, err
	}

	if err := plan.Load(); err != nil {
		return nil, err
	}

	fmt.Printf("%d targets found\n", len(plan.Targets))
	if len(plan.Targets) < 1 {
		return plan, nil
	}

	planPath, err := plan.Save()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Plan saved: %s\n", planPath)

	return plan, nil
}

func ExecutePlan(planName string) error {
	planPath := FindPlan(planName)
	plan, err := NewPlanFromPath(planPath)
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

	if err := plan.Execute(); err != nil {
		return err
	}

	return nil
}
