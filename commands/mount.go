package commands

import (
	"fmt"
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

var mountListCmd = &cobra.Command{
	Use:   "list <unit_name>",
	Short: "List bravetools mounts active on a Unit",
	Long:  "Shows active mounted disk devices managed by bravetools on a Unit",
	Run:   mountList,
}

func init() {
	mountDir.AddCommand(mountListCmd)
}

func mount(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) < 2 {
		log.Fatal("missing <source> UNIT:<target>")
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

func mountList(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		err := host.ListAllMounts()
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, arg := range args {
		fmt.Printf("Mounts for %s:\n", arg)
		err := host.ListMounts(arg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
