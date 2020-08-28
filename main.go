package main

import (
	"log"

	"github.com/beringresearch/bravetools/commands"
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
