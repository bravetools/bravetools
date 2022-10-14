package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var publishImageNames []string

var bravePublish = &cobra.Command{
	Use:   "publish <instance> [<instance>...]",
	Short: "Publish deployed Units as images",
	Long:  `Published Units will be saved in the current working directory as *.tar.gz file`,
	Run:   publish,
}

func init() {
	bravePublish.Flags().StringSliceVar(&publishImageNames, "image_name", []string{}, "Image names to apply to exported units")
}

func publish(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing name - please provide unit name")
	}

	for i, name := range args {
		imageName := ""
		if i < len(publishImageNames) {
			imageName = publishImageNames[i]
		}
		err := host.PublishUnit(name, imageName)
		if err != nil {
			log.Fatal(err)
		}
	}
}
