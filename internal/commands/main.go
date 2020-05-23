package commands

import (
	"reflect"
)

type Command interface {
	Run(target string) error
	Validate() error
}

func CommandStructFields(s interface{}) []Command {
	commands := []Command{}

	elem := reflect.ValueOf(s).Elem()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if !field.IsNil() {
			commands = append(commands, field.Interface().(Command))
		}
	}

	return commands
}
