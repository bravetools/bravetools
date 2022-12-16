package main

import (
	_ "embed"

	"log"

	"github.com/bravetools/bravetools/commands"
	"github.com/bravetools/bravetools/shared"
)

//go:embed VERSION
var versionString string

type cmdGlobal struct {
	configPath string
}

func main() {
	shared.Version = versionString

	err := commands.BravetoolsCmd.Execute()
	if err != nil && err.Error() != "" {
		log.Fatal(err)
	}
}
