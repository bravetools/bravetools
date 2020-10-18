package commands

import (
	"log"

	"github.com/bravetools/bravetools/platform"
	"github.com/spf13/cobra"
)

var configureHost = &cobra.Command{
	Use:   "configure",
	Short: "Configure local host parameters",
	Long:  `Update host configuration using settings in $HOME/.bravetools/config.yaml`,
	Run:   configure,
}

func configure(cmd *cobra.Command, args []string) {
	checkBackend()
	err := platform.ConfigureHost(host.Settings, host.Remote)
	if err != nil {
		log.Fatal(err)
	}
}
