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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return func() []string {
			var imageNames []string
			images, _ := platform.GetLocalImages()
			for _, image := range images {
				imageNames = append(imageNames, image.String())
			}
			return imageNames
		}(), cobra.ShellCompDirectiveNoFileComp
	},
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
