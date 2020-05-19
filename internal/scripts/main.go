package scripts

import (
	"fmt"
	"os"
	"path"

	"github.com/dropseed/workhorse/internal/config"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func Find(dir, name, extension string) string {
	if fileExists(name) {
		return name
	}
	return path.Join("workhorse", dir, fmt.Sprintf("%s.%s", name, extension))
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

func CreatePlan(scriptName, planName string) error {
	scriptPath := Find("scripts", scriptName, "yml")
	config, err := config.NewConfigFromPath(scriptPath)
	if err != nil {
		return err
	}

	plan, err := NewPlan(scriptPath, config)
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

	planPath, err := plan.Save(planName)
	if err != nil {
		return err
	}
	fmt.Printf("Plan saved: %s\n", planPath)

	return nil
}

func ExecutePlan(planName string) error {
	plan, err := NewPlanFromPath(Find("plans", planName, "json"))
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
