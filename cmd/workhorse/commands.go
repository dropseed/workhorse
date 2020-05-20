package main

import (
	"fmt"
	"reflect"

	"github.com/google/go-github/v31/github"
	"github.com/spf13/cobra"
)

var commands = &cobra.Command{
	Use:   "commands",
	Short: "Output pr commands by reflection",
	Run: func(cmd *cobra.Command, args []string) {
		client := github.NewClient(nil)
		// client.PullRequests.mer
		fooType := reflect.TypeOf(client.PullRequests)
		for i := 0; i < fooType.NumMethod(); i++ {
			method := fooType.Method(i)
			fmt.Println(method.Name)

			// s := reflect.ValueOf(method.Type.In(4)).Elem()
			for x := 5; x < method.Type.NumIn(); x++ {
				in := method.Type.In(x)
				fmt.Printf("%s\n", in)

				if in.Kind() == reflect.Ptr {
					println("POITNE")
				}
				if in.Kind() == reflect.Struct {
					println("STRU")
					for f := 0; f < in.NumField(); f++ {
						// Skip all NodeID fields
						fmt.Printf("  %s\n", in.Field(f))
					}
				}
			}
			fmt.Println()

			// s := reflect.ValueOf(&inputStruct).Elem()

			// typeOfT := s.Type()
			// for i := 0; i < s.NumField(); i++ {
			// 	f := s.Field(i)
			// 	fmt.Printf("%d: %s %s = %v\n", i,
			// 		typeOfT.Field(i).Name, f.Type())
			// }

			// fmt.Printf("%+v", inputStruct.Type().Name())
			// for x := 0; x < inputStruct.NumField(); x++ {
			// 	// 	fmt.Printf("%v\n", inputStruct.Field(x).Type().Name())
			// 	fmt.Printf("%v\n", reflect.Indirect(inputStruct.Field(x)).Type().Name())
			// }
			// println()

			// in := make([]reflect.Value, method.Type.NumIn())

			// for x := 0; x < method.Type.NumIn(); x++ {
			// 	t := method.Type.In(x)
			// 	// object := objects[t]
			// 	// fmt.Println(i, "->", object)
			// 	// in[i] = reflect.ValueOf(object)

			// 	fmt.Printf("%v %v\n", x, t)
			// }
		}
	},
}

func init() {
	rootCmd.AddCommand(commands)
}
