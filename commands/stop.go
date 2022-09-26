package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var braveStop = &cobra.Command{
	Use:   "stop [<remote>:]<instance> [[<remote>:]<instance>...]",
	Short: "Stop Units",
	Long:  ``,
	Run:   stop,
}

func stop(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing name - please provide unit name")
		return
	}

	for _, arg := range args {
		err := host.StopUnit(arg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
