package commands

import (
	"log"
	"strings"

	"github.com/bravetools/bravetools/platform"
	"github.com/bravetools/bravetools/shared"

	"github.com/spf13/cobra"
)

var baseBuild = &cobra.Command{
	Use:   "base DISTRIBUTION/RELEASE/ARCH",
	Short: "Pull a base image from LXD Image Server or public Github Bravefile",
	Long: `Import images available at "public" image server (default https://cloud-images.ubuntu.com/releases) or
from Bravefiles stored in public GitHub repositories`,
	Run: buildBase,
}

var remoteName string

func init() {
	includeBaseFlags(baseBuild)
}

func includeBaseFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&remoteName, "remote", "r", "local", "Name of the remote which will be used to build the base image.")
}

func buildBase(cmd *cobra.Command, args []string) {
	checkBackend()
	var err error

	if len(args) == 0 {
		log.Fatal("missing name - please provide a base name")
		return
	}

	if strings.HasPrefix(args[0], "github.com/") {
		bravefile, err = shared.GetBravefileFromGitHub(args[0])
		if err != nil {
			log.Fatal(err)
		}

	} else {
		bravefile, err = platform.GetBravefileFromLXD(args[0])
		if err != nil {
			log.Fatal(err)
		}
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
	if err != nil {
		log.Fatal(err)
	}
}
