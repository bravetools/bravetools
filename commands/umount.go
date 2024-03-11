package commands

import (
	"log"
	"strings"

	"github.com/bravetools/bravetools/platform"
	"github.com/spf13/cobra"
)

var umountDir = &cobra.Command{
	Use:   "umount UNIT:<path> [UNIT:<path>...]",
	Short: "Unmount device mounted on <path> from UNIT",
	Long:  ``,
	Run:   umount,
}

func umount(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing UNIT:<path>")
		return
	}

	//Mounts are supported only over a local remote
	remote, err := platform.LoadRemoteSettings("local")
	if err != nil {
		log.Fatal(err)
	}

	host.Remote = remote

	for _, arg := range args {
		remote := strings.SplitN(arg, ":", -1)
		if len(remote) == 1 {
			log.Fatal("target directory should be specified as UNIT:<path>")
		}

		err := host.UmountShare(remote[0], remote[1])
		if err != nil {
			log.Fatal(err)
		}
	}

}
