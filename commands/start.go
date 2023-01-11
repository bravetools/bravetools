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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return host.GetUnitNames(), cobra.ShellCompDirectiveNoFileComp
	},
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
