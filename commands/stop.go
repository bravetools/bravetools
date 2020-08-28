package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var braveStop = &cobra.Command{
	Use:   "stop NAME",
	Short: "Stop Unit",
	Long:  ``,
	Run:   stop,
}

func stop(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Missing name - please provide unit name")
		return
	}

	err := host.StopUnit(args[0], backend)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
