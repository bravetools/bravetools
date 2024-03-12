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
	cmd.Flags().StringVarP(&unitConfig, "config", "", "", "Path to the Bravefile file [OPTIONAL]")
	cmd.Flags().StringVarP(&remoteName, "remote", "r", "local", "Name of a Bravetools remote that will build the image")
	cmd.Flags().StringVarP(&bravefilePath, "path", "p", "", "Absolute path to directory containing the Bravefile [OPTIONAL]")
}

func build(cmd *cobra.Command, args []string) {
	p := "Bravefile"

	if bravefilePath != "" {
		p = path.Join(bravefilePath, "Bravefile")
	}
	if unitConfig != "" {
		p = unitConfig
	}

	err := bravefile.Load(p)

	if err != nil {
		log.Fatal("failed to load Bravefile: ", err)
	}

	remote, err := platform.LoadRemoteSettings(remoteName)
	if err != nil {
		log.Fatal(err)
	}

	host.Remote = remote

	if remote.Name != "local" {
		host.Settings.StoragePool.Name = remote.Storage
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
