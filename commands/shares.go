package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var braveShares = &cobra.Command{
	Use:   "shares [delete] <share_name>",
	Short: "Show/Delete shared data volumes",
	Long:  `shares shows shared data volumes in Brave storage pool and deletes them by name`,
	Run:   shares,
}

func shares(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 1 {
		fmt.Fprintln(os.Stderr, "Missing shared data volume name")
		return
	}

	if len(args) == 0 {
		err := host.ListVolumes()
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(args) == 2 {
		if args[0] != "delete" {
			fmt.Fprintln(os.Stderr, "Incorrect option. Did you mean delete?")
			return
		}
		err := host.DeleteVolume(args[1])
		if err != nil {
			log.Fatal(err)
		}
	}
}
