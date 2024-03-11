package commands

import (
	"log"
	"strings"

	"github.com/bravetools/bravetools/platform"
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

	remoteInfo := strings.SplitN(args[1], ":", -1)
	if len(remoteInfo) == 1 {
		log.Fatal("target directory should be specified as UNIT:<target>")
	}

	//Mounts are supported only over a local remote
	remote, err := platform.LoadRemoteSettings("local")
	if err != nil {
		log.Fatal(err)
	}

	host.Remote = remote

	err = host.MountShare(args[0], remoteInfo[0], remoteInfo[1])
	if err != nil {
		log.Fatal(err)
	}
}
