package commands

import (
	"fmt"
	"log"
	"os"
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
		fmt.Fprintln(os.Stderr, "Missing UNIT:<disk>")
		return
	}

	remote := strings.SplitN(args[0], ":", -1)
	if len(remote) == 1 {
		log.Fatal("Target directory should be specified as UNIT:<disk>")
	}

	err := host.UmountDirectory(remote[0], remote[1])
	if err != nil {
		log.Fatal(err)
	}

}
