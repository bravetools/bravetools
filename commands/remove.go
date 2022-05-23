package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var braveRemove = &cobra.Command{
	Use:   "remove NAME",
	Short: "Remove a Unit or an Image",
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

	if imageToggle {
		err := host.DeleteLocalImage(args[0])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := host.DeleteUnit(args[0])
		if err != nil {
			log.Fatal(err)
		}

		err = host.DeleteImageByName(args[0])
		if err != nil {
			log.Fatal(err)
		}
	}
}
