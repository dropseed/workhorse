package github

import (
	"errors"
	"fmt"
	"time"
)

type Sleep struct {
}

func (cmd *Sleep) Run(gh *GitHub, owner string, repo string, number int, args ...interface{}) error {
	arg := args[0]

	switch seconds := arg.(type) {
	default:
		return errors.New("Unknown type")
	case int:
		duration := time.Duration(int64(seconds))
		time.Sleep(duration * time.Second)
	case int64:
		duration := time.Duration(seconds)
		time.Sleep(duration * time.Second)
	case float64:
		duration := time.Duration(seconds)
		time.Sleep(duration * time.Second)
	}

	return nil
}

func (cmd *Sleep) Validate(args ...interface{}) error {
	if len(args) != 1 {
		return errors.New("Should have at exactly one number")
	}
	for _, arg := range args {
		switch v := arg.(type) {
		default:
			return fmt.Errorf("Arg is not a int: %v", v)
		case int:
			continue
		case int64:
			continue
		case float64:
			continue
		}
	}
	return nil
}
