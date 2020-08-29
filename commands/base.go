package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/beringresearch/bravetools/shared"

	"github.com/spf13/cobra"
)

var baseBuild = &cobra.Command{
	Use:   "base NAME",
	Short: "Build a base unit",
	Long: `Build a base unit from images available at https://images.linuxcontainers.org.
Command accepts image names in the format Distribution/Release/Architecture`,
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

	fmt.Println("Building base unit: ", bravefile.PlatformService.Name)

	err = host.BuildUnit(bravefile)
	if err != nil {
		log.Fatal(err)
	}
}
