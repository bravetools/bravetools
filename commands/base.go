package commands

import (
	"log"
	"strings"

	"github.com/bravetools/bravetools/shared"

	"github.com/spf13/cobra"
)

var baseBuild = &cobra.Command{
	Use:   "base DISTRIBUTION/RELEASE/ARCH",
	Short: "Pull a base image from LXD Image Server or public Github Bravefile",
	Long: `Import images available at https://images.linuxcontainers.org or
from Bravefiles stored in public GitHub repositories`,
	Run: buildBase,
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
		bravefile, err = shared.GetBravefileFromLXD(args[0])
		if err != nil {
			log.Fatal(err)
		}
	}

	err = host.BuildImage(bravefile)
	if err != nil {
		log.Fatal(err)
	}
}
