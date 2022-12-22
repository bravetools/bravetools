package commands

import (
	"log"

	"github.com/bravetools/bravetools/platform"
	"github.com/spf13/cobra"
)

var braveExportImage = &cobra.Command{
	Use:   "export <image> [<image>...]",
	Short: "Export bravetools image from local image store as LXD tarball",
	Long:  ``,
	Run:   exportImage,
}

var imageExportDir string

func init() {
	braveExportImage.Flags().StringVarP(&imageExportDir, "out", "o", "", "Directory to export images to [OPTIONAL]")
}

func exportImage(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing image name")
		return
	}

	for _, arg := range args {
		err := platform.ExportBravetoolsImage(arg, imageExportDir)
		if err != nil {
			log.Fatal(err)
		}
	}
}
