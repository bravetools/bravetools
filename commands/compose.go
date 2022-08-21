package commands

import (
	"log"
	"os"
	"path/filepath"

	"github.com/bravetools/bravetools/shared"
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

	baseDir := "."
	if len(args) > 0 {
		baseDir = args[0]
		_, err := os.Stat(baseDir)
		if err != nil {
			log.Fatal("unable to read brave-compose.yaml: ", err)
		}
	}

	// Load composefile from directory. Favour ".yaml" over ".yml" but accept both.
	if shared.FileExists(filepath.Join(baseDir, shared.ComposefileName)) {
		p = filepath.Join(baseDir, shared.ComposefileName)
	} else {
		if shared.FileExists(filepath.Join(baseDir, shared.ComposefileAlias)) {
			p = filepath.Join(baseDir, shared.ComposefileAlias)
		}
	}
	if p == "" {
		log.Fatalf("composefile %q not found at %q", shared.ComposefileName, baseDir)
	}

	err := composefile.Load(p)

	if err != nil {
		log.Fatal("failed to load compose file: ", err)
	}

	err = host.Compose(backend, composefile)
	if err != nil {
		log.Fatal(err)
	}
}
