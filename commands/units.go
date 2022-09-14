package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var braveListUnits = &cobra.Command{
	Use:   "units [NAME]",
	Short: "List Units",
	Long: `This function returns a list of all Units deployed on a remote Bravetools host. 

If a specific remote is not specified, all units across all remotes will be returned.`,
	Run:  units,
	Args: cobra.RangeArgs(0, 1),
}

func units(cmd *cobra.Command, args []string) {
	checkBackend()

	remoteName := ""
	if len(args) == 1 {
		remoteName = args[0]
	}

	err := host.ListUnits(backend, remoteName)
	if err != nil {
		log.Fatal(err)
	}
}
