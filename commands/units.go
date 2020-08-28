package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var braveListUnits = &cobra.Command{
	Use:   "units",
	Short: "List Units",
	Long:  `This function returns a list of all Units deployed on a remote Bravetools host`,
	Run:   units,
}

func units(cmd *cobra.Command, args []string) {
	checkBackend()
	err := host.ListUnits(backend)
	if err != nil {
		log.Fatal(err)
	}
}
