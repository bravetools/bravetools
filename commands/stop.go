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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return host.GetUnitNames(), cobra.ShellCompDirectiveNoFileComp
	},
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
