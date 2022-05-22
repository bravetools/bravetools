package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
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
		fmt.Fprintln(os.Stderr, "Missing name - please provide a base name")
		return
	}

	if strings.Contains(args[0], "github.com") {
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

	// Resource checks
	info, err := backend.Info()
	if err != nil {
		log.Fatal(err)
	}

	usedDiskSize, err := shared.SizeCountToInt(info.Disk[0])
	if err != nil {
		log.Fatal(err)
	}
	totalDiskSize, err := shared.SizeCountToInt(info.Disk[1])
	if err != nil {
		log.Fatal(err)
	}

	if (totalDiskSize - usedDiskSize) < 5000000000 {
		err = errors.New("not enough free storage on disk")
		log.Fatal(err)
	}

	err = host.BuildImage(bravefile)
	if err != nil {
		log.Fatal(err)
	}
}
