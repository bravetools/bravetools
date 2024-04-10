package commands

import (
	"log"
	"os"
	"path/filepath"

	"github.com/bravetools/bravetools/platform"
	"github.com/bravetools/bravetools/shared"
	"github.com/spf13/cobra"
)

var braveCompose = &cobra.Command{
	Use:   "compose",
	Short: "Compose a system from a set of images",
	Long:  ``,
	Run:   compose,
}

func init() {
	includeComposeFlags(braveCompose)
}

func includeComposeFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&remoteName, "remote", "r", "local", "Name of a Bravetools remote that will build the image")
}

func compose(cmd *cobra.Command, args []string) {
	var composefilePath string
	baseDir := "."

	// If user passes a path directly to composefile, set that
	if len(args) > 0 {
		stat, err := os.Stat(args[0])
		if err != nil {
			log.Fatalf("unable to find resolve path %q\n", args[0])
		}
		if stat.IsDir() {
			baseDir = args[0]
		} else {
			composefilePath = args[0]
		}
	}

	// If composefile not set, then user-provided path is a dir. Attempt to find composefile in dir.
	if composefilePath == "" {
		// Load composefile from directory. Favour ".yaml" over ".yml" but accept both.
		if shared.FileExists(filepath.Join(baseDir, shared.ComposefileName)) {
			composefilePath = filepath.Join(baseDir, shared.ComposefileName)
		} else {
			if shared.FileExists(filepath.Join(baseDir, shared.ComposefileAlias)) {
				composefilePath = filepath.Join(baseDir, shared.ComposefileAlias)
			}
		}
	}

	// If composefile path still not set it was not found - fail with err
	if composefilePath == "" {
		log.Fatalf("composefile %q not found at %q", shared.ComposefileName, baseDir)
	}

	err := composefile.Load(composefilePath)

	if err != nil {
		log.Fatal("failed to load compose file: ", err)
	}

	// Load build backend
	remote, err := platform.LoadRemoteSettings(remoteName)
	if err != nil {
		log.Fatal(err)
	}

	host.Remote = remote

	if remote.Name != "local" {
		host.Settings.StoragePool.Name = remote.Storage
	}

	err = host.Compose(backend, composefile)
	if err != nil {
		log.Fatal(err)
	}
}
