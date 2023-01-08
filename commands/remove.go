package commands

import (
	"log"

	"github.com/bravetools/bravetools/platform"
	"github.com/spf13/cobra"
)

var braveRemove = &cobra.Command{
	Use:   "remove [<remote>:]<instance> [[<remote>:]<instance>...]",
	Short: "Remove Units or Images",
	Long:  ``,
	Run:   remove,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		if imageToggle {
			return func() []string {
				var imageNames []string
				images, _ := platform.GetLocalImages()
				for _, image := range images {
					imageNames = append(imageNames, image.String())
				}
				return imageNames
			}(), cobra.ShellCompDirectiveNoFileComp
		}
		return host.GetUnitNames(), cobra.ShellCompDirectiveNoFileComp
	},
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
			err := host.DeleteLocalImage(arg, false)
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
