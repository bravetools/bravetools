package commands

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var braveStart = &cobra.Command{
	Use:   "start NAME",
	Short: "Start Unit",
	Long:  ``,
	Run:   start,
}

func start(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing name - please provide unit name")
	}

	err := host.StartUnit(args[0], backend)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
