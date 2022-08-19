package commands

import (
	"log"
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
		log.Fatal("Failed to load compose file: ", err)
	}

	err = host.Compose(backend, composefile)
	if err != nil {
		log.Fatal(err)
	}
}
