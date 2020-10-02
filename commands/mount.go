package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var mountDir = &cobra.Command{
	Use:   "mount [UNIT:]<source> UNIT:<target>",
	Short: "Mount a directory to a Unit",
	Long:  `mount supports local directories as well as shared directories between Units.`,
	Run:   mount,
}

func mount(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Missing <source> UNIT:<target>")
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
