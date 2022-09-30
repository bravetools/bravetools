package commands

import (
	"log"
	"path"

	"github.com/bravetools/bravetools/platform"
	"github.com/spf13/cobra"
)

var braveBuild = &cobra.Command{
	Use:   "build",
	Short: "Build an image from a Bravefile",
	Long:  ``,
	Run:   build,
}

var bravefilePath string

func init() {
	includePathFlags(braveBuild)
}

func includePathFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&bravefilePath, "path", "p", "", "Absolute path to Bravefile [OPTIONAL]")
}

func build(cmd *cobra.Command, args []string) {
	var p string

	if bravefilePath == "" {
		p = "Bravefile"
	} else {
		p = path.Join(bravefilePath, "Bravefile")
	}

	err := bravefile.Load(p)

	if err != nil {
		log.Fatal("Failed to load Bravefile: ", err)
	}

	err = host.BuildImage(*bravefile)
	switch errType := err.(type) {
	case nil:
	case *platform.ImageExistsError:
		log.Fatalf("image %q already exists - if you want to rebuild it, first delete the existing image with: `brave remove -i [IMAGE]`", errType.Name)
	default:
		log.Fatal(err)
	}
}
