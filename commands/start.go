package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var braveStart = &cobra.Command{
	Use:   "start  [<remote>:]<instance> [[<remote>:]<instance>...]",
	Short: "Start Units",
	Long:  ``,
	Run:   start,
}

func start(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing name - please provide unit name")
	}

	for _, arg := range args {
		err := host.StartUnit(arg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
