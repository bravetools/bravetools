package commands

import (
	"log"
	"path"

	"github.com/spf13/cobra"
)

var braveCompose = &cobra.Command{
	Use:   "compose",
	Short: "Compose a system from a set of images",
	Long:  ``,
	Run:   compose,
}

func compose(cmd *cobra.Command, args []string) {
	var p string

	if len(args) == 0 {
		p = "brave-compose.yml"
	} else {
		p = path.Join(args[0], "brave-compose.yml")
	}

	err := composefile.Load(p)

	if err != nil {
		log.Fatal("Failed to load compose file: ", err)
	}

	err = host.Compose(backend, composefile)
	if err != nil {
		log.Fatal(err)
	}
}
