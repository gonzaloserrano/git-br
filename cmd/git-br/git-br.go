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
	err := gitbr.Open(path)
	if err != nil {
		fmt.Sprintf("Error: %s", err.Error())
		os.Exit(1)
	}
	println("Goodbye!")
}
