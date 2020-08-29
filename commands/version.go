package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beringresearch/bravetools/shared"
)

var (
	braveVersion = &cobra.Command{
		Use:              "version",
		Short:            "Show current bravetools version",
		Long:             ``,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {},
		Run:              version,
	}
)

func version(cmd *cobra.Command, args []string) {
	fmt.Println(shared.VersionString())
}
