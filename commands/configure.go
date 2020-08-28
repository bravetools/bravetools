package commands

import (
	"log"

	"github.com/beringresearch/bravetools/platform"
	"github.com/spf13/cobra"
)

var configureHost = &cobra.Command{
	Use:   "configure",
	Short: "Configure local host parameters such as storage",
	Long:  `Bravetools reads configuration specifications from ~/.bravetools/config.yaml and configures host accordingly`,
	Run:   configure,
}

func configure(cmd *cobra.Command, args []string) {
	checkBackend()
	err := platform.ConfigureHost(host.Settings, host.Remote)
	if err != nil {
		log.Fatal(err)
	}
}
