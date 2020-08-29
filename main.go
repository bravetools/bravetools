package main

import (
	"log"

	"github.com/bravetools/bravetools/commands"
)

type cmdGlobal struct {
	configPath string
}

func main() {

	err := commands.BravetoolsCmd.Execute()
	if err != nil && err.Error() != "" {
		log.Fatal(err)
	}
}
