package main

import (
	"fmt"
	"os"

	gitbr "github.com/gonzaloserrano/git-br"
)

func main() {
	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	uiRunner, err := gitbr.Open(path)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	err = uiRunner.Run()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	println("Goodbye!")
}
