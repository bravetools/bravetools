package commands

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var braveStop = &cobra.Command{
	Use:   "stop [<remote>:]<instance>",
	Short: "Stop Unit",
	Long:  ``,
	Run:   stop,
}

func stop(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing name - please provide unit name")
		return
	}

	err := host.StopUnit(args[0])
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
