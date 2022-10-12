package commands

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var mountDir = &cobra.Command{
	Use:   "mount [UNIT:]<source> UNIT:<target>",
	Short: "Mount a directory to a Unit",
	Long:  `mount local directories as well as shared volumes between Units.`,
	Run:   mount,
}

func mount(cmd *cobra.Command, args []string) {
	checkBackend()

	if len(args) == 0 {
		err := host.ListAllMounts()
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if len(args) == 1 {
		err := host.ListMounts(args[0])
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	remote := strings.SplitN(args[1], ":", -1)
	if len(remote) == 1 {
		log.Fatal("Target directory should be specified as UNIT:<target>")
	}

	err := host.MountShare(args[0], remote[0], remote[1])
	if err != nil {
		log.Fatal(err)
	}
}
