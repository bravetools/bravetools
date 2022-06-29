package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var braveRemove = &cobra.Command{
	Use:   "remove [<remote>:]<instance> [[<remote>:]<instance>...]",
	Short: "Remove Units or Images",
	Long:  ``,
	Run:   remove,
}
var imageToggle bool

func init() {
	includeRemoveFlags(braveRemove)
}

func includeRemoveFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&imageToggle, "image", "i", false, "Toggle to delete a local image")
}

func remove(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing name - please provide unit name")
		return
	}

	for _, arg := range args {
		if imageToggle {
			err := host.DeleteLocalImage(arg)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err := host.DeleteUnit(arg)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
