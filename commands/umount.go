package commands

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var umountDir = &cobra.Command{
	Use:   "umount UNIT:<disk>",
	Short: "Unmount <disk> from UNIT",
	Long:  ``,
	Run:   umount,
}

func umount(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing UNIT:<disk>")
		return
	}

	for _, arg := range args {
		remote := strings.SplitN(arg, ":", -1)
		if len(remote) == 1 {
			log.Fatal("target directory should be specified as UNIT:<disk>")
		}

		err := host.UmountShare(remote[0], remote[1])
		if err != nil {
			log.Fatal(err)
		}
	}

}
