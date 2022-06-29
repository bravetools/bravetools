package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var braveImportImage = &cobra.Command{
	Use:   "import <file> [<file>...]",
	Short: "Import LXD image tarballs into local Bravetools image repository",
	Long:  ``,
	Run:   importImage,
}

func importImage(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing name - please provide tarball name")
		return
	}

	for _, arg := range args {
		err := host.ImportLocalImage(arg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
