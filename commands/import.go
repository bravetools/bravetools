package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var braveImportImage = &cobra.Command{
	Use:   "import NAME",
	Short: "Import an LXD image tarball into local Bravetools image repository",
	Long:  ``,
	Run:   importImage,
}

func importImage(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Missing name - please provide tarball name")
		return
	}

	err := host.ImportLocalImage(args[0])
	if err != nil {
		log.Fatal(err)
	}
}
