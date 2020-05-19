package github

import (
	"errors"
	"time"
)

type Sleep struct {
}

func (cmd *Sleep) Run(gh *GitHub, owner string, repo string, number int, args ...interface{}) error {
	seconds := args[0].(int)
	duration := time.Duration(int64(seconds))
	time.Sleep(duration * time.Second)
	return nil
}

func (cmd *Sleep) Validate(args ...interface{}) error {
	if len(args) != 1 {
		return errors.New("Should have at exactly one number")
	}
	for _, arg := range args {
		if _, ok := arg.(int); !ok {
			return errors.New("Arg is not a int")
		}
	}
	return nil
}
