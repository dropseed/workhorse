package main

import (
	"fmt"
	"os"
)

func printErrAndExitFailure(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		printErrAndExitFailure(err)
	}
}
